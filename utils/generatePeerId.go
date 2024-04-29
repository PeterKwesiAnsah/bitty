package utils

import (
	"crypto/rand"
	"fmt"
)

func GeneratePeerId() ([]byte, error) {
	peerIdLenth := 20
	buf := make([]byte, peerIdLenth)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("generating peerId failed:%w", err)
	}
	return buf, nil
}
