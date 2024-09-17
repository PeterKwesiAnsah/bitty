package worker

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/peterkwesiansah/bitty/bitfield"
	"github.com/peterkwesiansah/bitty/message"
)

type WorkPiece struct {
	PieceIndex int
	PieceHash  [20]byte
	Length     int
}

const (
	//maximum number of bytes requested from a peer
	maxBufSize     = 16384
	indexLength    = 4
	beginLength    = 4
	headerLength   = indexLength + beginLength
	maxActivePeers = 5
)

type Worker struct {
	Conn net.Conn
	Bf   bitfield.Bitfield
}

func (wkr *Worker) CheckPieceIntegrity(downloadedPiece []byte, pieceHash [20]byte) bool {
	buf := downloadedPiece
	//check file piece integrity
	hash := sha1.New()
	hash.Write(buf)
	bufHash := hash.Sum(nil)

	return bytes.Equal(bufHash, pieceHash[:])
}

// keeps track of the downloading of a piece
type downloadManager struct {
	pieceIndex         int
	pieceBuf           []byte
	maxUnfulfilledReqs int
	requested          int
	startDownload      bool
	requests           int
	downloaded         int
}

// SendUnchoke tells the peer you are ready to talk
func (wkr *Worker) SendUnchoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := wkr.Conn.Write(msg.Serialize())
	return err
}

// Interested sends an Interested message to the peer
func (wkr *Worker) Interested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := wkr.Conn.Write(msg.Serialize())
	return err
}

// requests buffer from peer
func (wkr *Worker) request(requested, blockSize, index int) error {
	payloadLength := 12
	reqBuf := make([]byte, payloadLength)
	offset := 4

	binary.BigEndian.PutUint32(reqBuf[0:offset], uint32(index))
	binary.BigEndian.PutUint32(reqBuf[offset:offset*2], uint32(requested))
	binary.BigEndian.PutUint32(reqBuf[offset*2:offset*3], uint32(blockSize))

	reqMsg := message.Message{ID: message.MsgRequest, Payload: reqBuf}

	_, err := wkr.Conn.Write(reqMsg.Serialize())

	if err != nil {
		return fmt.Errorf("blocksize request failed %w", err)
	}

	return nil
}

// read,process incoming request messages
func (wkr *Worker) processMsgPiece(dm *downloadManager, msg message.Message) error {

	if len(msg.Payload) < headerLength {
		return fmt.Errorf("piece message too small")
	}

	pieceIndex := int(binary.BigEndian.Uint32(msg.Payload[0:indexLength]))
	if pieceIndex != dm.pieceIndex {
		return fmt.Errorf("expected pieceIndex %d but got this pieceIndex %d", dm.pieceIndex, pieceIndex)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[indexLength:headerLength]))
	pieceBuf := msg.Payload[headerLength:]
	if begin+len(pieceBuf) > len(dm.pieceBuf) {
		return fmt.Errorf("offset %d is too large", begin)
	}

	copy(dm.pieceBuf[begin:], pieceBuf)
	dm.downloaded = dm.downloaded + len(pieceBuf)

	dm.requests = dm.requests - 1
	return nil
}

func (wkr *Worker) processMsg(dm *downloadManager) error {
	msg, err := message.ReadMessage(wkr.Conn)

	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}
	msgId := msg.ID

	if msgId == message.MsgChoke {
		dm.startDownload = false
	} else if msgId == message.MsgUnchoke {
		dm.startDownload = true
	} else if msgId == message.MsgPiece {
		return wkr.processMsgPiece(dm, *msg)
	}
	return nil

}

func (wkr *Worker) Download(pieceIndex, pieceLength int) ([]byte, error) {

	dm := &downloadManager{maxUnfulfilledReqs: 5, startDownload: true, pieceIndex: pieceIndex, pieceBuf: make([]byte, pieceLength)}

	wkr.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer wkr.Conn.SetDeadline(time.Time{})

	for dm.downloaded < pieceLength {
		if dm.startDownload {
			for dm.requested < pieceLength && dm.requests < dm.maxUnfulfilledReqs {
				blockSize := min(pieceLength-dm.requested, maxBufSize)
				//send request to peer
				err := wkr.request(dm.requested, blockSize, dm.pieceIndex)
				if err != nil {
					return nil, err
				}
				dm.requests++
				dm.requested = dm.requested + blockSize
			}
		}

		//read and parse message from peer
		err := wkr.processMsg(dm)
		if err != nil {
			return nil, fmt.Errorf("failed to download %w", err)
		}
	}
	return dm.pieceBuf, nil
}
