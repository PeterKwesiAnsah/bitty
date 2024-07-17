package worker

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/peterkwesiansah/bitty/client"
	"github.com/peterkwesiansah/bitty/message"
)

type workPiece struct {
	pieceIndex int
	pieceHash  [20]byte
	length     int
}

const (
	//maximum number of bytes you can ask from a peer
	maxBufSize     = 16384
	indexLength    = 4
	beginLength    = 4
	headerLength   = indexLength + beginLength
	maxActivePeers = 5
)

type worker struct {
	client         *client.Client
	wp             workPiece
	pieceBuf       []byte
	pieceIntegrity bool
}

func (wkr *worker) request(requested, blockSize int) error {
	payloadLength := 12
	reqBuf := make([]byte, payloadLength)
	offset := 4

	binary.BigEndian.PutUint32(reqBuf[0:offset], uint32(wkr.wp.pieceIndex))
	binary.BigEndian.PutUint32(reqBuf[offset:offset*2], uint32(requested))
	binary.BigEndian.PutUint32(reqBuf[offset*2:offset*3], uint32(blockSize))

	reqMsg := message.Message{ID: message.MsgPiece, Payload: reqBuf}

	_, err := wkr.client.Conn.Write(reqMsg.Serialize())
	if err != nil {
		return fmt.Errorf("blocksize request failed %w", err)
	}

	return nil
}

func (wkr *worker) checkPieceIntegrity() bool {

	return false

}

type downloadManager struct {
	maxUnfulfilledReqs int
	requested          int
	startDownload      bool
	requests           int
}

func (wkr *worker) processMsg(dm *downloadManager) error {
	msg, err := message.ReadMessage(wkr.client.Conn)

	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}
	msgId := msg.ID

	return nil

}

func (wkr *worker) download() ([]byte, error) {

	dm := downloadManager{maxUnfulfilledReqs: 5}

	wkr.client.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer wkr.client.Conn.SetDeadline(time.Time{})

	for dm.requested < wkr.wp.length {
		if dm.startDownload {
			for dm.requested < wkr.wp.length && dm.requests < dm.maxUnfulfilledReqs {
				blockSize := min(wkr.wp.length-dm.requested, maxBufSize)
				//send request to peer
				err := wkr.request(dm.requested, blockSize)
				if err != nil {
					return nil, err
				}
				dm.requests++
				dm.requested = dm.requested + blockSize
			}
			//read and parse message from peer
		}

	}
	return nil, nil
}
