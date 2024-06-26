package torrent

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"math"
	"net"
	"os"
	"sync"

	"github.com/codecrafters-io/bittorrent-starter-go/helpers"
)

var (
	SIXTEEN_KILO_BYTES = 16 * 1024
)

type Bittorrent struct {
	*Torrent
	*TrackerResponse
	NumberOfBlocks int
}

func NewBittorrent(torrentFile string) (*Bittorrent, error) {

	torrent, err := NewTorrent(torrentFile)

	if err != nil {
		return nil, err
	}

	tracker, err := NewTracker(*torrent)

	if err != nil {
		panic(err)
	}

	response, err := tracker.Get()

	if err != nil {
		panic(err)
	}

	return &Bittorrent{
		Torrent:         torrent,
		TrackerResponse: &response,
		NumberOfBlocks:  0,
	}, nil
}

func (b *Bittorrent) Handshake(peer string) (*Communication, error) {

	var handshake bytes.Buffer

	tcpServer, err := net.ResolveTCPAddr("tcp", peer)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp", nil, tcpServer)

	if err != nil {
		return nil, err
	}

	handshake.WriteByte(byte(19))
	handshake.WriteString("BitTorrent protocol")
	handshake.Write(make([]byte, 8))
	handshake.Write(b.Torrent.Hash)
	handshake.WriteString(helpers.RandomPeerId())

	_, err = conn.Write(handshake.Bytes())

	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 1024)

	size, err := conn.Read(buffer)
	if err != nil {

		return nil, err
	}

	return &Communication{
		Conn:   conn,
		PeerId: hex.EncodeToString(buffer[48:size]),
		Ip:     peer,
	}, nil
}

func (b *Bittorrent) generatesBlocks() map[int][]*Block {

	var sum, index int

	b.NumberOfBlocks = int(math.Ceil(float64(b.Torrent.Info.PieceLength) / float64(SIXTEEN_KILO_BYTES)))

	blocks := make(map[int][]*Block)

	for i := range b.Torrent.piecesHash() {
		blocks[i] = make([]*Block, 0)
		for j := 0; j < b.NumberOfBlocks; j++ {
			block := &Block{
				lengthPrefix: 13,
				id:           6,
				index:        uint32(i),
				begin:        uint32(j * int(SIXTEEN_KILO_BYTES)),
				length:       uint32(SIXTEEN_KILO_BYTES),
			}

			sum += SIXTEEN_KILO_BYTES

			if sum > b.Torrent.Info.Length {
				block.length = uint32(b.Torrent.Info.Length - (sum - SIXTEEN_KILO_BYTES))
				blocks[i] = append(blocks[i], block)
				break
			}
			blocks[i] = append(blocks[i], block)

			index++
		}
	}

	return blocks

}

func (b Bittorrent) DownloadAPiece(path string, index int) error {

	// For one piece use the first peer
	firstPeer := b.TrackerResponse.Peers[0].String()

	communication, err := b.Handshake(firstPeer)

	if err != nil {
		return err
	}

	err = communication.sendInteresting()

	if err != nil {
		return err
	}

	blocks := b.generatesBlocks()

	bytes, err := communication.Download(blocks[index])

	if err != nil {
		return err
	}

	err = helpers.CompareHashes(index, b.Torrent.piecesHash()[index], sha1.Sum(bytes))

	if err != nil {
		return err
	}

	err = os.WriteFile(path, bytes, 777)

	if err != nil {
		return err
	}

	log.Printf("Piece %d downloaded to /tmp/test-piece-%d.", index, index)

	return nil
}

func (b Bittorrent) DownloadLow(output string) error {

	// For one piece use the first peer
	firstPeer := b.TrackerResponse.Peers[0].String()

	communication, err := b.Handshake(firstPeer)

	if err != nil {
		return err
	}

	err = communication.sendInteresting()

	if err != nil {
		return err
	}

	blocks := b.generatesBlocks()

	data := make([]byte, 0)

	for index := 0; index < len(blocks); index++ {

		res, err := communication.Download(blocks[index])

		err = helpers.CompareHashes(index, b.Torrent.piecesHash()[index], sha1.Sum(res))

		if err != nil {
			return err
		}

		data = append(data, res...)
	}

	_ = os.Remove(output)

	err = os.WriteFile(output, data, 777)

	if err != nil {
		return err
	}

	log.Printf("Downloaded %s to %s.", b.Torrent.Info.Name, output)

	return nil
}

func (b Bittorrent) initCommunications() ([]*Communication, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	size := len(b.TrackerResponse.Peers)
	communications := make([]*Communication, size)
	tracker := make(chan *Communication)
	trackerErr := make(chan error)
	index := 0

	for _, peer := range b.TrackerResponse.Peers {
		go func(peer IP, ctx context.Context) {

			select {
			case <-ctx.Done():
				return
			default:
				communication, err := b.Handshake(peer.String())

				if err != nil {
					trackerErr <- err
					cancel()
					return
				}

				err = communication.sendInteresting()

				if err != nil {
					trackerErr <- err
					cancel()
					return
				}

				tracker <- communication
			}

		}(peer, ctx)

	}

	for {
		select {
		case communication := <-tracker:
			communications[index] = communication
			index++
			if index == size {
				close(tracker)
				return communications, nil
			}
		case err := <-trackerErr:

			for _, communication := range communications {
				if communication != nil {
					communication.Close()
				}
			}

			return nil, err
		}
	}

}

func (b Bittorrent) closeCommunications(communications []*Communication) error {

	var err error

	for _, communication := range communications {
		err = communication.Close()

		if err != nil {
			return err
		}
	}

	return nil

}

func (b Bittorrent) DownloadFast(output string) error {

	var wg sync.WaitGroup

	communications, err := b.initCommunications()

	if err != nil {
		return err
	}

	//erros := make(chan error)
	blocks := b.generatesBlocks()
	queue := helpers.NewQueue(communications...)

	contents := make([][]byte, len(blocks))

	for index := 0; index < len(blocks); index++ {

		wg.Add(1)
		go func(blocks []*Block, index int) {

			defer wg.Done()

			communication, _ := queue.Pop()

			data, err := (*communication).Download(blocks)

			contents[index] = data

			if err != nil {
				log.Fatal(err)
				//erros <- err
			}

			queue.Add(*communication)

		}(blocks[index], index)
	}

	wg.Wait()

	file, err := os.OpenFile(output, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	for _, content := range contents {

		_, err = file.Write(content)

		if err != nil {
			return err
		}
	}

	defer b.closeCommunications(communications)

	return nil
	/*



		fmt.Println(len(data))

		err = os.WriteFile(output, data, 777)

		if err != nil {
			return err
		}

		log.Printf("Downloaded %s to %s.", b.Torrent.Info.Name, output)

		return nil
	*/
}
