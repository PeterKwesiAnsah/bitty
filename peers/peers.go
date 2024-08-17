package peers

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jackpal/bencode-go"
	"github.com/peterkwesiansah/bitty/bencodeTorrent"
)

type Peer struct {
	IPAddress net.IP
	Port      uint16
}

type peersResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type peersTracker struct {
	url    string
	peerId []byte
}

type p2p struct {
	Peers  []Peer
	PeerId []byte
}

func trackerInfo(infoHash [20]byte, announce string, infoLength int) (*peersTracker, error) {
	const port uint16 = 6881

	base, err := url.Parse(announce)
	if err != nil {
		return nil, err
	}

	const peerIdLength = 20
	peerIdBuf := make([]byte, peerIdLength)
	n, err := rand.Read(peerIdBuf)
	if err != nil {
		return nil, fmt.Errorf("generating peerId failed:%w", err)
	}
	if n != 20 {
		return nil, fmt.Errorf("generate peer ID: insufficient random data")
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerIdBuf)},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(infoLength)},
	}
	base.RawQuery = params.Encode()
	return &peersTracker{
		url:    base.String(),
		peerId: peerIdBuf,
	}, nil
}

func parsePeers(peers []byte) ([]Peer, error) {
	const (
		peerLength = 6
		portLength = 2
	)
	peersCount := len(peers) / peerLength
	if len(peers)%peerLength != 0 {
		return nil, fmt.Errorf("received corrupted data")
	}
	peersBuf := make([]Peer, peersCount)
	for i := 0; i < peersCount; i++ {
		endOffSet := (i + 1) * peerLength
		peersBuf[i].IPAddress = net.IP(peers[i*peerLength : endOffSet-portLength])
		peersBuf[i].Port = binary.BigEndian.Uint16([]byte(peers[endOffSet-portLength : endOffSet]))
	}
	return peersBuf, nil
}

// findPeers sends an HTTP GET request to the tracker with the specified parameters
func Peers(bct *bencodeTorrent.BencodeTorrent) (*p2p, error) {
	pt, err := trackerInfo(bct.InfoHash, bct.Announce, bct.Info.Length)
	if err != nil {
		return nil, err
	}
	// Send the HTTP GET request to the tracker
	resp, err := http.Get(pt.url)
	if err != nil {
		return nil, fmt.Errorf("GET request to tracker url failed %w", err)
	}

	defer resp.Body.Close()
	pr := peersResp{}

	err = bencode.Unmarshal(resp.Body, &pr)

	if err != nil {
		return nil, fmt.Errorf("reading peers failed %w", err)
	}
	peerslist, err := parsePeers([]byte(pr.Peers))
	if err != nil {
		return nil, err
	}
	return &p2p{
		Peers:  peerslist,
		PeerId: pt.peerId,
	}, nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IPAddress.String(), strconv.Itoa(int(p.Port)))
}
