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

type HandShake struct {
	PeerID   [20]byte
	InfoHash [20]byte
	Protocol string //Bitorrent Protocol
}

func ReadBinHandshakeToStruct(r io.Reader) (*HandShake, error) {
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
	return &HandShake{PeerID: [20]byte(peerID), InfoHash: [20]byte(infoHash), Protocol: protocolStr}, nil

}

func (hs *HandShake) BinhandShake() []byte {
	hsBuf := make([]byte, len(hs.InfoHash)+len(hs.PeerID)+len(hs.Protocol)+lenOfUnusedBuf+lenToStoreProtocol)
	hsBuf[0] = byte(len(hs.Protocol))
	cur := lenToStoreProtocol
	cur = cur + copy(hsBuf[cur:len(hs.Protocol)+lenToStoreProtocol], []byte(hs.Protocol))
	cur = cur + lenOfUnusedBuf
	cur = cur + copy(hsBuf[cur:cur+lenOfInfoHash], hs.InfoHash[:])
	copy(hsBuf[cur:], hs.PeerID[:])
	return hsBuf
}
