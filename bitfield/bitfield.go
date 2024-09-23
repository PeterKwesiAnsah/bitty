package bitfield

import (
	"fmt"
	"io"

	"github.com/peterkwesiansah/bitty/message"
)

// A Bitfield represents the pieces that a peer has
type Bitfield []byte

// HasPiece tells if a bitfield has a particular index set
func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}
	return bf[byteIndex]>>uint(7-offset)&1 != 0
}

func Read(conn io.Reader) (Bitfield, error) {

	msg, err := message.ReadMessage(conn)
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
