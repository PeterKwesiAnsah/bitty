package torrentFile

import (
	"log"
	"github.com/jackpal/bencode-go"
	"github.com/peterkwesiansah/bitty/utils"
)

type torrentFileInfo struct {
	Length      int64
	Name        string
	PieceLength int64
	Pieces      string
}

type TorrentFile struct {
	Announce     string
	CreatedBy    string
	CreationDate int64
	Encoding     string
	Info         torrentFileInfo
}

func Decode(pathToTorrent string) TorrentFile {
	tf := TorrentFile{}
	torrentFileDescriptor, err := utils.ReadTorrentFile(pathToTorrent)
	errInDecode := bencode.Unmarshal(torrentFileDescriptor, &tf)
	defer torrentFileDescriptor.Close()
	if err != nil || errInDecode != nil {
		log.Fatalln(err.Error(), errInDecode.Error())
	}
	return tf
}
