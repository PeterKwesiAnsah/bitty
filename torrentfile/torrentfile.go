package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/peterkwesiansah/bitty/client"
	"github.com/peterkwesiansah/bitty/message"
	"github.com/peterkwesiansah/bitty/peers"
)

type filePiece struct {
	PieceIndex int
	PieceHash  [20]byte
	Length     int
}
type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type downloadPieceState struct {
	index           int
	client          *client.Client
	pieceSize       int
	requestsCount   int
	requestedBytes  int
	bytesDownloaded int
	pieceData       []byte
}

// This is the maximum number of bytes you can ask from a peer
const (
	MaxBufSize   = 16384
	indexLength  = 4
	beginLength  = 4
	headerLength = indexLength + beginLength
)

var messageIDtoFunc = map[message.MessageID]func(ps *downloadPieceState, msg message.Message) error{
	message.MsgUnchoke: handleMsgUnchoke,
	message.MsgChoke:   handleMsgChoke,
	message.MsgPiece:   handleMsgPiece,
	message.MsgHave:    handleMsgHave,
}

func sendDataRequest(c *client.Client, filePieceIndex, requestedBytes, bufSizeToRequest int) error {
	messagePayload := make([]byte, 12)

	binary.BigEndian.PutUint32(messagePayload[0:4], uint32(filePieceIndex))
	binary.BigEndian.PutUint32(messagePayload[4:8], uint32(requestedBytes))
	binary.BigEndian.PutUint32(messagePayload[8:12], uint32(bufSizeToRequest))

	requestMessage := message.Message{
		ID:      message.MsgRequest,
		Payload: messagePayload,
	}
	_, err := c.Conn.Write(requestMessage.BinMessage())

	return err
}

func handleMsgUnchoke(ps *downloadPieceState, msg message.Message) error {
	ps.client.Choked = false
	return nil
}
func handleMsgChoke(ps *downloadPieceState, msg message.Message) error {
	ps.client.Choked = true
	return nil
}
func handleMsgPiece(ps *downloadPieceState, msg message.Message) error {
	//log.Println("handling message piece")
	//receivedMessage, _ := message.ReadBinMessageToStruct(ps.client.Conn)

	if len(msg.Payload) < headerLength {
		return fmt.Errorf("piece message too small")
	}
	filePieceIndex := int(binary.BigEndian.Uint32(msg.Payload[0:indexLength]))
	if filePieceIndex != ps.index {
		return fmt.Errorf("expected pieceIndex %d but got this pieceIndex %d", ps.index, filePieceIndex)
	}
	begin := int(binary.BigEndian.Uint32(msg.Payload[indexLength:headerLength]))
	pieceBuf := msg.Payload[headerLength:]
	if begin+len(pieceBuf) > ps.pieceSize {
		return fmt.Errorf("offset %d is too large", begin)
	}
	copy(ps.pieceData[begin:], pieceBuf)
	ps.bytesDownloaded = ps.bytesDownloaded + len(pieceBuf)
	ps.requestsCount = ps.requestsCount - 1
	return nil
}

func handleMsgHave(ps *downloadPieceState, msg message.Message) error {
	//msg, err := message.ReadBinMessageToStruct(ps.client.Conn)
	pieceIndexLength := 4
	if len(msg.Payload) < 4 {
		return fmt.Errorf("message payload size too small")
	}
	pieceIndexBuf := msg.Payload[0:pieceIndexLength]
	pieceIndex := binary.BigEndian.Uint32(pieceIndexBuf)
	ps.client.Bitfield.SetPiece(int(pieceIndex))

	return nil

}

func (ps *downloadPieceState) handlePeerResponse() error {
	//This should take at most 30 secs, else read timeout triggers
	receivedMessage, err := message.ReadBinMessageToStruct(ps.client.Conn)
	if err != nil {
		return fmt.Errorf("failed to read peer message %w", err)
	}
	if receivedMessage == nil {
		return nil
	}

	if handleMessage, ok := messageIDtoFunc[message.MessageID(receivedMessage.ID)]; ok {
		return handleMessage(ps, *receivedMessage)
	} else {
		return fmt.Errorf("unexpected messageId %d", int(receivedMessage.ID))
	}
}

func (ps *downloadPieceState) verifyIntegrity(pieceHash [20]byte) error {
	// Create a new SHA-1 hash
	hash := sha1.New()

	// Write the data to the hash
	hash.Write(ps.pieceData)

	hashInBytes := hash.Sum(nil)

	if bytes.Equal(hashInBytes, pieceHash[:]) {
		return nil
	}

	return fmt.Errorf("piece integrity check failed")

}

func downloadPiece(c *client.Client, fp *filePiece) (*downloadPieceState, error) {
	state := downloadPieceState{
		index:     fp.PieceIndex,
		client:    c,
		pieceSize: fp.Length,
		pieceData: make([]byte, fp.Length),
	}
	//log.Printf("attempting to download piece index %d", fp.PieceIndex)
	// prevent us from being stuck on unresponsive peer
	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{}) // Disable the deadline

	for state.bytesDownloaded < state.pieceSize {
		//log.Println(state.bytesDownloaded, state.pieceSize, fp.PieceIndex)
		if !state.client.Choked {
			//send a queue of requests
			for state.requestedBytes < state.pieceSize && state.requestsCount < 5 {
				bufSizeToRequest := min(state.pieceSize-state.requestedBytes, MaxBufSize)
				err := sendDataRequest(c, fp.PieceIndex, state.requestedBytes, bufSizeToRequest)
				if err != nil {
					return nil, err
				}
				state.requestsCount = state.requestsCount + 1
				state.requestedBytes = state.requestedBytes + bufSizeToRequest
			}
		}
		// process each request in order FIFO
		error := state.handlePeerResponse()
		if error != nil {
			return nil, fmt.Errorf("failed to download piece %w", error)
		}
	}
	log.Panicln("hurray")
	// log.Printf("piece index %d downloaded successfully", fp.PieceIndex)
	// log.Printf("checking integrity of downloaded piece %d ...", fp.PieceIndex)
	// //integerity of downloaded piece is checked here
	// error := state.verifyIntegrity(fp.PieceHash)
	// if error != nil {
	// 	return nil, fmt.Errorf("piece %d integrity verification failed %w", fp.PieceIndex, error)
	// }
	// log.Printf("piece %d integrity verified successfully", fp.PieceIndex)
	// c.SendHave(fp.PieceIndex)
	return &state, nil
}

func (t *Torrent) startDownloadingFilePieces(p peers.Peer, filePiecesQueue chan *filePiece, downloadedPiecesQueue chan downloadPieceState) {

	c, err := client.Connect(p, t.PeerID, t.InfoHash)
	if err != nil {
		//replace peer (have a channel of peers ??)
		log.Printf("failed to connect to peer:%v", err)
		return
	}
	log.Printf("connection successful for peer %v", p)
	c.SendUnchoke()
	c.Interested()

	for fp := range filePiecesQueue {
		if !c.Bitfield.HasPiece(fp.PieceIndex) {
			filePiecesQueue <- fp
			continue
		}
		log.Printf("Attempting to downloading piece %d", fp.PieceIndex)

		downloadedPiece, err := downloadPiece(c, fp)
		if err != nil {
			//log.Println(fp.Length)
			log.Printf("failed to download a piece %v", err)
			filePiecesQueue <- fp
			continue
		}
		downloadedPiecesQueue <- *downloadedPiece
		log.Printf("piece %d result sent to channel", fp.PieceIndex)
	}
}

func (t *Torrent) Download() ([]byte, error) {
	//creating a buffered chanel to hold len(pieceHashes) before blocking
	piecesToBeDownloadedQueue := make(chan *filePiece, len(t.PieceHashes))
	piecesDownloadedQueue := make(chan downloadPieceState)
	bytesDownloaded := 0
	torrentFileDownloaded := make([]byte, t.Length)
	//feeding the buffered chanel up to len(pieceHashes)
	for i, pieceHash := range t.PieceHashes {
		piecesToBeDownloadedQueue <- &filePiece{
			PieceIndex: i,
			PieceHash:  pieceHash,
			Length:     t.PieceLength,
		}
	}

	for _, peer := range t.Peers {
		//log.Printf("spewing off worker %d", index)
		//errors
		go t.startDownloadingFilePieces(peer, piecesToBeDownloadedQueue, piecesDownloadedQueue)
	}
	for bytesDownloaded < t.Length {
		downloadedPiece := <-piecesDownloadedQueue
		start := downloadedPiece.index * downloadedPiece.pieceSize
		// end := start + downloadedPiece.pieceSize
		// if end > t.Length {
		// 	end = t.Length
		// }
		bytesDownloaded = bytesDownloaded + downloadedPiece.bytesDownloaded
		fmt.Printf("piece index %d downloaded successfully ", downloadedPiece.index)
		copy(torrentFileDownloaded[start:], downloadedPiece.pieceData)

	}
	close(piecesDownloadedQueue)
	close(piecesToBeDownloadedQueue)
	return torrentFileDownloaded, nil

}
