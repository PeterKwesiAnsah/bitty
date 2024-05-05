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

type messageID uint8

type message struct {
	ID      messageID
	payload []byte
}

func ReadBinMessageToStruct(r io.Reader) (*message, error) {

	messagebuf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading message stream:%w", err)
	}

	messageID := messageID(messagebuf[sizeOfLength])
	payload := messagebuf[sizeOfLength+1:]
	return &message{ID: messageID, payload: payload}, nil
}

func (m *message) BinMessage() []byte {
	totalSizeOfMessage := sizeOfLength + sizeOfId + len(m.payload)
	messageBuf := make([]byte, totalSizeOfMessage)
	lengthBuf := messageBuf[:sizeOfLength]
	binary.BigEndian.PutUint32(lengthBuf, uint32(totalSizeOfMessage))
	copy(messageBuf[:sizeOfLength], lengthBuf)
	messageBuf[sizeOfLength] = byte(m.ID)
	copy(messageBuf[sizeOfLength+1:], m.payload)
	return messageBuf

}
