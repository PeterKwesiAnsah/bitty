package torrentFile

import (
	"bytes"
	"crypto/sha1"
	"github.com/jackpal/bencode-go"
	"github.com/peterkwesiansah/bitty/utils"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	CreatedBy    string      `bencode:"created by"`
	CreationDate int64       `bencode:"creation date"`
	Encoding     string      `bencode:"encoding"`
	Info         bencodeInfo `bencode:"info"`
}

// Serializes the bencodeInfo into a SHA-1 hash
func (bct *bencodeInfo) GetInfoHash() ([20]byte, error) {
	var buf bytes.Buffer
	errInSerializing := bencode.Marshal(&buf, *bct)
	if errInSerializing != nil {
		return [20]byte{}, errInSerializing
	}
	return sha1.Sum(buf.Bytes()), nil
}

/*
*
returns a slice of an array of bytes , with each array representing the SHA-1 hash of a piece
*/
func (bct *bencodeInfo) GetPieceHashes() [][20]byte {
	hashLen := 20
	piecesInBytes := []byte(bct.Pieces)
	numOfPieceHashes := len([]byte(bct.Pieces)) / hashLen
	pieceHashes := make([][20]byte, numOfPieceHashes)
	endCopyIndex := 20
	for i := 0; i < numOfPieceHashes; i++ {
		startCopyIndex := hashLen * i
		endCopyIndex = copy(pieceHashes[i][:], piecesInBytes[startCopyIndex:endCopyIndex]) + endCopyIndex
	}
	return pieceHashes
}

func Decode(pathToTorrent string) (bencodeTorrent, error) {
	bct := bencodeTorrent{}
	torrentFileDescriptor, err := utils.ReadTorrentFile(pathToTorrent)
	if err != nil {
		return bct, err
	}

	defer torrentFileDescriptor.Close()

	errInDecode := bencode.Unmarshal(torrentFileDescriptor, &bct)
	if errInDecode != nil {
		return bencodeTorrent{}, errInDecode
	}
	return bct, nil
}
