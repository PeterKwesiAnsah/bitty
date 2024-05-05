package handShake

import (
	"fmt"
	"io"
)

const (
	lenToStoreProtocol = 1
	lenOfProtocol      = 13
	lenOfPeerID        = 20
	lenOfInfoHash      = 20
	lenOfUnusedBuf     = 8
)

type handShake struct {
	peerID   [20]byte
	infoHash [20]byte
	protocol string //Bitorrent Protocol
}

func ReadBinHandshakeToStruct(r io.Reader) (*handShake, error) {
	handShakeBuf, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading handshake bin stream into buff:%w", err)
	}
	protocoLen := int(handShakeBuf[0])
	protocolStr := string(handShakeBuf[lenToStoreProtocol : protocoLen+lenToStoreProtocol])
	if protocoLen != len(protocolStr) {
		return nil, fmt.Errorf("invalid protocol")
	}
	startInfoHashBufIndex := lenToStoreProtocol + protocoLen + lenOfUnusedBuf
	infoHash := handShakeBuf[startInfoHashBufIndex : startInfoHashBufIndex+lenOfInfoHash]
	peerID := handShakeBuf[startInfoHashBufIndex+lenOfInfoHash:]
	return &handShake{peerID: [20]byte(peerID), infoHash: [20]byte(infoHash), protocol: protocolStr}, nil

}

func (hs *handShake) BinhandShake() []byte {
	hsBuf := make([]byte, len(hs.infoHash)+len(hs.peerID)+len(hs.protocol)+lenOfUnusedBuf+lenToStoreProtocol)
	hsBuf[0] = byte(len(hs.protocol))
	cur := lenToStoreProtocol
	cur = cur + copy(hsBuf[cur:len(hs.protocol)+lenToStoreProtocol], []byte(hs.protocol))
	cur = cur + lenOfUnusedBuf
	cur = cur + copy(hsBuf[cur:cur+lenOfInfoHash], hs.infoHash[:])
	copy(hsBuf[cur:], hs.peerID[:])
	return hsBuf
}
