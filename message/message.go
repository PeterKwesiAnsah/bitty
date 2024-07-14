package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	//Every message has a length buffer which contains 32 bits unsigned integer,requiring 4 bytes
	sizeOfLength = 4
	//Because is an 8 bits unsigned Integer
	sizeOfId = 1
)

const (
	// MsgChoke chokes the receiver
	MsgChoke MessageID = 0
	// MsgUnchoke unchokes the receiver
	MsgUnchoke MessageID = 1
	// MsgInterested expresses interest in receiving data
	MsgInterested MessageID = 2
	// MsgNotInterested expresses disinterest in receiving data
	MsgNotInterested MessageID = 3
	// MsgHave alerts the receiver that the sender has downloaded a piece
	MsgHave MessageID = 4
	// MsgBitfield encodes which pieces that the sender has downloaded
	MsgBitfield MessageID = 5
	// MsgRequest requests a block of data from the receiver
	MsgRequest MessageID = 6
	// MsgPiece delivers a block of data to fulfill a request
	MsgPiece MessageID = 7
	// MsgCancel cancels a request
	MsgCancel MessageID = 8
)

type MessageID uint8

type Message struct {
	ID      MessageID
	Payload []byte
}

func ReadBinMessageToStruct(r io.Reader) (*Message, error) {

	const messageIdlen = 4
	messageIDbuf := make([]byte, int(messageIdlen))
	_, err := io.ReadFull(r, messageIDbuf)

	if err != nil {
		return nil, fmt.Errorf("parsing message stream %w", err)
	}
	messageSize := binary.BigEndian.Uint32(messageIDbuf)
	// keep-alive message
	if messageSize == 0 {
		return nil, nil
	}
	messageBuf := make([]byte, int(messageSize))

	_, err = io.ReadFull(r, messageBuf)

	if err != nil {
		return nil, fmt.Errorf("parsing message stream %w", err)
	}

	messageID := MessageID(messageBuf[0])
	payload := messageBuf[1:]

	return &Message{ID: messageID, Payload: payload}, nil
}

func (m *Message) BinMessage() []byte {
	if m == nil {
		return make([]byte, 4)
	}
	totalSizeOfMessage := sizeOfLength + sizeOfId + len(m.Payload)
	messageBuf := make([]byte, totalSizeOfMessage)
	lengthBuf := make([]byte, sizeOfLength)

	binary.BigEndian.PutUint32(lengthBuf, uint32(sizeOfId+len(m.Payload)))
	copy(messageBuf[:sizeOfLength], lengthBuf)

	messageBuf[sizeOfLength] = byte(m.ID)
	copy(messageBuf[sizeOfLength+sizeOfId:], m.Payload)
	return messageBuf

}
