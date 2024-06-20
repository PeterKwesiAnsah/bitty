package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/peterkwesiansah/bitty/bitfield"
	handShake "github.com/peterkwesiansah/bitty/handshake"
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

func completeHandshake(conn net.Conn, infohash, peerID [20]byte) (*handShake.HandShake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{}) // Disable the deadline

	sendHandShake := &handShake.HandShake{
		PeerID:   peerID,
		InfoHash: infohash,
		Protocol: "BiTorrent protocol",
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
		return nil, fmt.Errorf("Hey man , you got the wrong file that i need")
	}

	return receivedHandShake, nil
}

// connect function makes a TCP connection with a peer with a time out of 3 seconds
func Connect(peer peers.Peer, peerID, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Connecting to peer failed:%w", err)
	}
	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		return nil, fmt.Errorf("Handshaking with peer failed:%w", err)
	}
	defer conn.Close()
	return nil, err

}
