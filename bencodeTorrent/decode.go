package bencodeTorrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"

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
	Announce     string `bencode:"announce"`
	CreatedBy    string `bencode:"created by"`
	CreationDate int64  `bencode:"creation date"`
	Encoding     string `bencode:"encoding"`
	Port         uint16 `bencode:"port"`

	Info bencodeInfo `bencode:"info"`
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

// findPeers sends an HTTP GET request to the tracker with the specified parameters
// and returns the response body or an error if one occurred.
func (bct *bencodeTorrent) FindPeers(peer_id string, port string) ([]byte, error) {
	infoHash, err := bct.Info.GetInfoHash()
	if err != nil {
		return nil, fmt.Errorf("get bencodeTorrent struct info hash: %w",err)
	}
	queryParams := url.Values{}
	queryParams.Set("info_hash", string(infoHash[:]))
	queryParams.Set("peer_id", peer_id)
	queryParams.Set("port", port)
	queryParams.Set("uploaded", "0")
	queryParams.Set("downloaded", "0")
	queryParams.Set("left", "100")
	queryParams.Set("compact", "1")
	// Construct the URL with query parameters
	fullURL := bct.Announce + "?" + queryParams.Encode()

	// Send the HTTP GET request to the tracker
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("error announcing to tracker: %v", err)
	}

	defer resp.Body.Close()

	// Read and return the response body
	//body := make([]byte, 1024) // Assuming response body won't exceed 1024 bytes
	//n, err := resp.Body.Read(body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	return body, nil
}

/*
*
returns a slice of an array of bytes , with each array representing the SHA-1 hash of a piece
*/
func (bci *bencodeInfo) GetPieceHashes() ([][20]byte, error) {
	hashLen := 20
	piecesInBytes := []byte(bci.Pieces)
	numOfPieceHashes := len([]byte(bci.Pieces)) / hashLen
	pieceHashes := make([][20]byte, numOfPieceHashes)
	if len([]byte(bci.Pieces))%hashLen != 0 {
		return nil, fmt.Errorf("received malformed pieces of length %d", len(piecesInBytes))
	}
	endCopyIndex := 20
	for i := 0; i < numOfPieceHashes; i++ {
		startCopyIndex := hashLen * i
		endCopyIndex = copy(pieceHashes[i][:], piecesInBytes[startCopyIndex:endCopyIndex]) + endCopyIndex
	}
	return pieceHashes, nil
}

// Serializes a bencode stream to a bencodeTorrent Struct
func Decode(pathToTorrent string) (bencodeTorrent, error) {
	bct := bencodeTorrent{}
	torrentFileDescriptor, err := utils.ReadTorrentFile(pathToTorrent)
	if err != nil {
		return bct, fmt.Errorf("reading torrent file:%w",err)
	}

	defer torrentFileDescriptor.Close()

	errInDecode := bencode.Unmarshal(torrentFileDescriptor, &bct)
	if errInDecode != nil {
		return bencodeTorrent{}, fmt.Errorf("decoding bencode stream failed: %w",errInDecode)
	}
	return bct, nil
}
