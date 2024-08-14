package bencodeTorrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce     string `bencode:"announce"`
	CreatedBy    string `bencode:"created by"`
	CreationDate int64  `bencode:"creation date"`
	Encoding     string `bencode:"encoding"`
	Port         uint16 `bencode:"port"`

	Info        bencodeInfo `bencode:"info"`
	InfoHash    [20]byte
	PieceHashes [][20]byte
}

// Serializes the bencodeInfo into a SHA-1 hash
func (bct *bencodeInfo) infoHash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *bct)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}

/*
*
returns a slice of an array of bytes , with each array representing the SHA-1 hash of a piece
*/
func (bci *bencodeInfo) pieceHashes() ([][20]byte, error) {
	const hashLength = 20
	piecesBuf := []byte(bci.Pieces)
	piecesBufLength := len(piecesBuf)
	pieceHashCount := piecesBufLength / hashLength
	pieceHashes := make([][hashLength]byte, pieceHashCount)

	if piecesBufLength%hashLength != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", piecesBufLength)
	}

	for i := 0; i < pieceHashCount; i++ {
		copy(pieceHashes[i][:], piecesBuf[i*hashLength:(i+1)*hashLength])
	}
	return pieceHashes, nil
}

// Takes a path to torrent ,read it into a bencode stream and serializes it to a bencodeTorrent Struct
func Decode(pathToTorrent string) (*bencodeTorrent, error) {
	bct := &bencodeTorrent{}
	file, err := os.Open(pathToTorrent)
	if err != nil {
		return nil, fmt.Errorf("reading torrent file %w", err)
	}

	defer file.Close()

	err = bencode.Unmarshal(file, bct)
	if err != nil {
		return nil, fmt.Errorf("decoding bencode stream failed: %w", err)
	}

	infoHash, err := bct.Info.infoHash()
	if err != nil {
		return nil, err
	}
	bct.InfoHash = infoHash

	pieceHashes, err := bct.Info.pieceHashes()
	if err != nil {
		return nil, err
	}
	bct.PieceHashes = pieceHashes

	return bct, nil
}
