package peers

import (
	"encoding/binary"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net"
	"net/http"
	"strconv"
)

type Peer struct {
	IPAddress net.IP
	Port      uint16
}

type trackerURLResStruct struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func BinPeersToPeer(peers []byte) ([]Peer, error) {
	lenOfPeer := 6
	lenOfPort := 2
	numOfPeers := len(peers) / lenOfPeer
	if len(peers)%lenOfPeer != 0 {
		//fmt.Println("hello")
		return nil, fmt.Errorf("received corrupted data")
	}
	totalPeersBuf := make([]Peer, numOfPeers)
	for i := 0; i < numOfPeers; i++ {
		endOfPeerIndex := (i + 1) * lenOfPeer
		//fmt.Println(endOfPeerIndex)
		totalPeersBuf[i].IPAddress = net.IP(peers[i*lenOfPeer : endOfPeerIndex-lenOfPort])
		totalPeersBuf[i].Port = binary.BigEndian.Uint16([]byte(peers[endOfPeerIndex-lenOfPort : endOfPeerIndex]))
	}
	return totalPeersBuf, nil
}

// findPeers sends an HTTP GET request to the tracker with the specified parameters
// and returns the response body or an error if one occurred.
func FindPeers(trackerURL string) ([]byte, error) {
	// Send the HTTP GET request to the tracker
	resp, err := http.Get(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("error announcing to tracker: %v", err)
	}

	defer resp.Body.Close()

	trackerURLResStruct := trackerURLResStruct{}

	err = bencode.Unmarshal(resp.Body, &trackerURLResStruct)

	if err != nil {
		return nil, fmt.Errorf("error decoding bencode stream response: %v", err)
	}
	return []byte(trackerURLResStruct.Peers), nil
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IPAddress.String(), strconv.Itoa(int(p.Port)))
}
