package torrentfile

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/peterkwesiansah/bitty/bitfield"
	handShake "github.com/peterkwesiansah/bitty/handshake"
	"github.com/peterkwesiansah/bitty/message"
	"github.com/peterkwesiansah/bitty/peers"
	"github.com/peterkwesiansah/bitty/worker"
)

type resultPiece struct {
	pieceIndex int
	pieceHash  [20]byte
	length     int
	buf        []byte
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

func (t *Torrent) performHandShake(conn net.Conn) error {
	hs := &handShake.HandShake{
		PeerID:   t.PeerID,
		InfoHash: t.InfoHash,
		Protocol: "BitTorrent protocol",
	}

	bufHs := hs.Serialize()
	_, err := conn.Write(bufHs)
	if err != nil {
		return err
	}

	readHs, err := handShake.Read(conn)
	if err != nil {
		return err
	}

	if !bytes.Equal(hs.InfoHash[:], readHs.InfoHash[:]) {
		return fmt.Errorf("wrong file")
	}
	return nil
}

func (t *Torrent) initDownloadWorker(peer peers.Peer, resultsQueue chan<- resultPiece, workQueue chan *worker.WorkPiece) error {

	//connect to peer
	conn, err := net.DialTimeout("tcp", peer.String(), 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer %w", err)
	}

	//set read and write connection timeout
	conn.SetDeadline(time.Now().Add(10 * time.Minute))
	defer func() {
		conn.SetDeadline(time.Time{})
		conn.Close()
	}()

	// handshake peer
	err = t.performHandShake(conn)
	if err != nil {
		return fmt.Errorf("failed to complete handshake with peer %w", err)
	}

	//receive bitfield
	bf, err := bitfield.Read(conn)
	if err != nil {
		return fmt.Errorf("failed to receive bitfield from peer %w", err)
	}

	//worker creation
	wkr := &worker.Worker{
		Conn: conn,
		Bf:   bf,
	}

	err = wkr.Interested()
	if err != nil {
		return fmt.Errorf("failed to send interested message %w", err)
	}

	err = wkr.SendUnchoke()
	if err != nil {
		return fmt.Errorf("failed to send unchoke message %w", err)
	}

	msg, err := message.ReadMessage(wkr.Conn)
	if err != nil {
		log.Print(err)
		return err
	}

	if msg == nil {
		return nil
	}

	if msg.ID != message.MsgUnchoke {
		return err
	}

	for wp := range workQueue {
		if !wkr.Bf.HasPiece(wp.PieceIndex) {
			//Try another workpiece
			workQueue <- wp
			continue
		}

		buf, err := wkr.Download(wp.PieceIndex, wp.Length)

		if err != nil {
			//Try another workpiece
			workQueue <- wp
			continue
		}

		if !wkr.CheckPieceIntegrity(buf, wp.PieceHash) {
			//Try another workpiece
			workQueue <- wp
			continue
		}

		//send to piece to  resultPiece chan
		rp := resultPiece{
			pieceIndex: wp.PieceIndex,
			buf:        buf,
			length:     wp.Length,
			pieceHash:  wp.PieceHash,
		}
		resultsQueue <- rp
	}
	return nil
}

func (t *Torrent) Download() ([]byte, error) {

	activePeersCount := len(t.Peers)
	donePieces := 0

	activePeers := make(chan peers.Peer, activePeersCount)
	errors := make(chan error, activePeersCount)
	workQueue := make(chan *worker.WorkPiece, len(t.PieceHashes))
	resultsQueue := make(chan resultPiece, len(t.PieceHashes))
	downloadCompleted := make(chan bool)

	fileBuf := make([]byte, t.Length)

	//creating queue of workPieces
	for index, pieceHash := range t.PieceHashes {
		workQueue <- &worker.WorkPiece{PieceIndex: index, PieceHash: pieceHash, Length: t.PieceLength}
	}

	for _, activePeer := range t.Peers {
		go func(peer peers.Peer) {

			//returns when there's nothing more to download in other words workQueue is empty or error
			err := t.initDownloadWorker(peer, resultsQueue, workQueue)
			errors <- err

		}(activePeer)
	}

	go func() {
		<-downloadCompleted
		//download is completed close all channels
		log.Println("Download completed, closing all channels")
		close(resultsQueue)
		close(workQueue)
		close(activePeers)
	}()

	go func() {
		for error := range errors {
			//handle error from connection related issues with peers
			if error != nil {
				log.Println(error.Error())
			}
		}

	}()

	for resultPiece := range resultsQueue {
		begin := resultPiece.pieceIndex * resultPiece.length
		copy(fileBuf[begin:], resultPiece.buf)

		donePieces++

		if donePieces == len(t.PieceHashes) {
			downloadCompleted <- true
		}

		percentComplete := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		log.Printf("%.2f%% complete\n", percentComplete)

	}
	return fileBuf, nil

}
