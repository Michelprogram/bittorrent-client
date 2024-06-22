package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/jackpal/bencode-go"
)

var (
	HASH_LEN           = 20
	SIXTEEN_KILO_BYTES = 16 * 1024
)

type Metafile struct {
	Announce string   `bencode:"announce"`
	Info     Metainfo `bencode:"info"`
}
type Metainfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type Torrent struct {
	Metafile
	Hash []byte
}

func (t Torrent) piecesHash() [][20]byte {

	piecesBuffer := []byte(t.Info.Pieces)
	size := len(piecesBuffer) / HASH_LEN

	hashes := make([][20]byte, size)

	for i := 0; i < size; i++ {
		hashes[i] = [20]byte(piecesBuffer[i*HASH_LEN : (i+1)*HASH_LEN])
	}

	return hashes
}

func (t Torrent) String() string {

	var hashesString string

	for i, hash := range t.piecesHash() {

		if i == len(t.piecesHash()) {
			hashesString += fmt.Sprintf("%x", hash)
		} else {
			hashesString += fmt.Sprintf("%x\n", hash)
		}

	}

	return fmt.Sprintf("Tracker URL: %s\nLength: %d\nInfo Hash: %x\nPiece Length: %d\nPiece Hashes:\n%s", t.Announce, t.Info.Length, t.Hash, t.Info.PieceLength, hashesString)
}

type Bittorrent struct {
	net.Conn
	torrent        *Torrent
	PeerId         string
	NumberOfBlocks int
}

func NewBittorrent() *Bittorrent {
	return &Bittorrent{
		Conn:           nil,
		torrent:        nil,
		PeerId:         "",
		NumberOfBlocks: 0,
	}
}

func (b Bittorrent) Info(path string) (Torrent, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return Torrent{}, err
	}

	info, err := NewBencode[map[string]any](string(data))

	if err != nil {
		return Torrent{}, nil
	}
	metafile := Metafile{
		Announce: info.Decoded["announce"].(string),
		Info: Metainfo{
			Length:      info.Decoded["info"].(map[string]any)["length"].(int),
			Name:        info.Decoded["info"].(map[string]any)["name"].(string),
			PieceLength: info.Decoded["info"].(map[string]any)["piece length"].(int),
			Pieces:      info.Decoded["info"].(map[string]any)["pieces"].(string),
		},
	}

	h := sha1.New()
	bencode.Marshal(h, metafile.Info)

	return Torrent{
		Metafile: metafile,
		Hash:     h.Sum(nil),
	}, nil

}

func (b Bittorrent) Receive(data string) (Bencode[any], error) {

	return NewBencode[any](data)
}

func (b *Bittorrent) Handshake(peer string, torrent Torrent) error {

	var handshake bytes.Buffer

	tcpServer, err := net.ResolveTCPAddr("tcp", peer)
	if err != nil {
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpServer)

	if err != nil {
		return err
	}

	handshake.WriteByte(byte(19))
	handshake.WriteString("BitTorrent protocol")
	handshake.Write(make([]byte, 8))
	handshake.Write(torrent.Hash)
	handshake.Write([]byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9})

	_, err = conn.Write(handshake.Bytes())

	if err != nil {
		return err
	}

	buffer := make([]byte, 1024)

	size, err := conn.Read(buffer)
	if err != nil {

		return err
	}

	b.PeerId = hex.EncodeToString(buffer[48:size])
	b.torrent = &torrent
	b.Conn = conn

	return nil
}

func (b Bittorrent) sendInteresting() error {

	_, err := b.Write([]byte{0, 0, 0, 1, 2})

	if err != nil {
		return err
	}

	err = Wait(b.Conn, 1)

	if err != nil {
		return errors.New("not a unchoke messageback")
	}

	return nil
}

func (b *Bittorrent) generatesBlocks() []*Block {

	var sum, index int

	b.NumberOfBlocks = b.torrent.Info.PieceLength / SIXTEEN_KILO_BYTES

	blocks := make([]*Block, b.NumberOfBlocks*len(b.torrent.piecesHash()))

	for i := range b.torrent.piecesHash() {
		for j := 0; j < b.NumberOfBlocks; j++ {
			block := &Block{
				lengthPrefix: 13,
				id:           6,
				index:        uint32(i),
				begin:        uint32(j * int(SIXTEEN_KILO_BYTES)),
				length:       uint32(SIXTEEN_KILO_BYTES),
			}

			sum += SIXTEEN_KILO_BYTES

			if i == len(b.torrent.piecesHash())-1 && j == b.NumberOfBlocks-1 {
				block.length = uint32(b.torrent.Info.Length - (sum - SIXTEEN_KILO_BYTES))
			}
			blocks[index] = block
			index++
		}
	}

	return blocks

}

func (b Bittorrent) compareHashes(index int, hashes [20]byte) error {

	if !bytes.Equal(b.torrent.piecesHash()[index][:], hashes[:]) {
		return fmt.Errorf("hash doesn't match at index %d : \nPiece hash : %x\nDownloaded hash :%x\n", index, b.torrent.piecesHash()[index], hashes)
	}

	return nil

}

func (b Bittorrent) downloadPiece(index uint32, blocks []*Block) []byte {

	var res []*Block
	var wg sync.WaitGroup

	for _, block := range blocks {
		if block.index == index {
			res = append(res, block)
			wg.Add(1)
			go func() {
				defer wg.Done()
				block.Request(b.Conn)
			}()

		}
	}

	wg.Wait()

	return res[0].Merge(res[1:])

}

func (b Bittorrent) Download(path string, index int) error {

	defer b.Close()

	if b.torrent == nil {
		return errors.New("handshake doesn't applied")
	}

	err := b.sendInteresting()

	if err != nil {
		return err
	}

	blocks := b.generatesBlocks()

	res := b.downloadPiece(uint32(index), blocks)

	err = b.compareHashes(index, sha1.Sum(res))

	if err != nil {
		return err
	}

	err = os.WriteFile(path, res, 666)

	if err != nil {
		return err
	}

	log.Println("Piece 0 downloaded to /tmp/test-piece-0.")

	return nil
}
