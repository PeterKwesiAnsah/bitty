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
	const (
		protocolLen  = 19
		handshakeLen = 68
	)

	handshakeBuf := make([]byte, handshakeLen)
	_, err := io.ReadFull(r, handshakeBuf)
	if err != nil {
		return nil, fmt.Errorf("reading handshake: %w", err)
	}

	if int(handshakeBuf[0]) != protocolLen {
		return nil, fmt.Errorf("invalid protocol length: got %d, want %d", handshakeBuf[0], protocolLen)
	}

	protocol := string(handshakeBuf[1:20])
	if protocol != "BitTorrent protocol" {
		return nil, fmt.Errorf("invalid protocol: %s", protocol)
	}

	var infoHash, peerID [20]byte
	copy(infoHash[:], handshakeBuf[28:48])
	copy(peerID[:], handshakeBuf[48:68])

	return &HandShake{
		Protocol: protocol,
		InfoHash: infoHash,
		PeerID:   peerID,
	}, nil
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
