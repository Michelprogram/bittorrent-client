package torrent

import (
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"

	"github.com/codecrafters-io/bittorrent-starter-go/helpers"
	"github.com/codecrafters-io/bittorrent-starter-go/internal/bencode"
)

var (
	SIZE_IP = 6
)

type IP struct {
	net.IP
	Port int
}

func (ip IP) String() string {
	return fmt.Sprintf("%v:%d", ip.IP, ip.Port)
}

type TrackerResponse struct {
	Interval int
	Peers    []IP
}

func NewTrackerResponse(interval int, peers []byte) *TrackerResponse {

	size := len(peers) / SIZE_IP
	ips := make([]IP, size)

	for i := 0; i < size; i++ {

		end := SIZE_IP*i + 6
		start := SIZE_IP * i

		ips[i] = IP{
			IP:   net.IPv4(peers[start:end][0], peers[start:end][1], peers[start:end][2], peers[start:end][3]),
			Port: int(big.NewInt(0).SetBytes(peers[start:end][4:]).Uint64()),
		}
	}

	return &TrackerResponse{
		Interval: interval,
		Peers:    ips,
	}
}

type Tracker struct {
	InfoHash   string
	PeerId     string
	Port       int
	Uploaded   int
	Downloaded int
	Left       int
	Compact    int
	*url.URL
}

func NewTracker(torrent Torrent) (*Tracker, error) {

	req, err := url.Parse(torrent.Announce)

	if err != nil {
		return nil, nil
	}

	return &Tracker{
		InfoHash:   string(torrent.Hash),
		PeerId:     helpers.RandomPeerId(),
		Port:       6881,
		Uploaded:   0,
		Downloaded: 0,
		Left:       torrent.Info.Length,
		Compact:    1,
		URL:        req,
	}, nil
}

func (t Tracker) Get() (TrackerResponse, error) {

	params := url.Values{}
	params.Add("info_hash", t.InfoHash)
	params.Add("peer_id", t.PeerId)
	params.Add("port", fmt.Sprintf("%d", t.Port))
	params.Add("uploaded", fmt.Sprintf("%d", t.Uploaded))
	params.Add("downloaded", fmt.Sprintf("%d", t.Downloaded))
	params.Add("left", fmt.Sprintf("%d", t.Left))
	params.Add("compact", fmt.Sprintf("%d", t.Compact))

	t.URL.RawQuery = params.Encode()

	res, err := http.Get(t.URL.String())
	if err != nil {
		return TrackerResponse{}, err
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)

	received, err := bencode.NewBencode[map[string]any](string(data))

	if err != nil {
		return TrackerResponse{}, err
	}

	interval := received.Decoded["interval"].(int)

	peers := received.Decoded["peers"].(string)

	response := NewTrackerResponse(interval, []byte(peers))

	return *response, err

}
