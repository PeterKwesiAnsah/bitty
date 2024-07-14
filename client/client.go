package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/peterkwesiansah/bitty/bitfield"
	handShake "github.com/peterkwesiansah/bitty/handshake"
	"github.com/peterkwesiansah/bitty/message"
	"github.com/peterkwesiansah/bitty/peers"
)

// A Client represents a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func recvBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	msg, err := message.ReadBinMessageToStruct(conn)
	if err != nil {
		return nil, err
	}

	if msg == nil {
		err := fmt.Errorf("expected bitfield but got %+v", msg)
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("expected bitfield but got ID %d", msg.ID)
		return nil, err
	}

	return msg.Payload, nil
}

func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handShake.HandShake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	sendHandShake := &handShake.HandShake{
		PeerID:   peerID,
		InfoHash: infohash,
		Protocol: "BitTorrent protocol",
	}

	sendHandShakeBin := sendHandShake.BinhandShake()

	// Hello, I am (peerID) do you talk (protocol) and do you have (infoHash)
	_, err := conn.Write(sendHandShakeBin)
	if err != nil {
		return nil, err
	}

	receivedHandShake, err := handShake.ReadBinHandshakeToStruct(conn)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(sendHandShake.InfoHash[:], receivedHandShake.InfoHash[:]) {
		return nil, fmt.Errorf("hey man , you got the wrong file that i need")
	}

	return receivedHandShake, nil
}

// connect function makes a TCP connection with a peer with a time out of 3 seconds, completes handshake and retrieves bf payload from peer
func Connect(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {

	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connecting to peer failed %w", err)
	}
	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshaking with peer failed %w", err)
	}

	//after successful handshake expect bitfield
	bitfield, err := recvBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("receiving bitfield from peer failed %w", err)
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bitfield,
		peer:     peer,
		peerID:   peerID,
		infoHash: infoHash,
	}, nil
}

// SendUnchoke sends an Unchoke message to the peer
func (c *Client) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.BinMessage())
	return err
}

// Interested sends an Interested message to the peer
func (c *Client) Interested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.BinMessage())
	return err
}

// SendHave sends a Have message to the peer
func (c *Client) SendHave(index int) error {
	indexLength := 4
	indexBuf := make([]byte, indexLength)
	binary.BigEndian.PutUint32(indexBuf, uint32(index))
	msg := message.Message{
		ID:      message.MsgHave,
		Payload: indexBuf,
	}
	_, err := c.Conn.Write(msg.BinMessage())
	return err
}
