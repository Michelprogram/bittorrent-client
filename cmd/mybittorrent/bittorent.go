package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"

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

func (b *Bittorrent) generatesBlocks() map[int][]*Block {

	var sum, index int

	b.NumberOfBlocks = int(math.Ceil(float64(b.torrent.Info.PieceLength) / float64(SIXTEEN_KILO_BYTES)))

	blocks := make(map[int][]*Block)

	for i := range b.torrent.piecesHash() {
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

			if sum > b.torrent.Info.Length {
				block.length = uint32(b.torrent.Info.Length - (sum - SIXTEEN_KILO_BYTES))
				blocks[i] = append(blocks[i], block)
				break
			}
			blocks[i] = append(blocks[i], block)

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

func (b Bittorrent) downloadPiece(blocks []*Block) []byte {

	var res []*Block

	for _, block := range blocks {
		res = append(res, block)
		_ = block.Request(b.Conn)
	}

	return res[0].Merge(res[1:])

}

func (b Bittorrent) DownloadPiece(path string, index int) error {

	defer b.Close()

	if b.torrent == nil {
		return errors.New("handshake doesn't applied")
	}

	err := b.sendInteresting()

	if err != nil {
		return err
	}

	blocks := b.generatesBlocks()

	res := b.downloadPiece(blocks[index])

	err = b.compareHashes(index, sha1.Sum(res))

	if err != nil {
		return err
	}

	err = os.WriteFile(path, res, 777)

	if err != nil {
		return err
	}

	log.Printf("Piece %d downloaded to /tmp/test-piece-%d.", index, index)

	return nil
}

func (b Bittorrent) DownloadWholePieces(output string) error {

	defer b.Close()

	if b.torrent == nil {
		return errors.New("handshake doesn't applied")
	}

	err := b.sendInteresting()

	if err != nil {
		return err
	}

	blocks := b.generatesBlocks()

	data := make([]byte, 0)

	for index := 0; index < len(blocks); index++ {

		fmt.Println(index)
		res := b.downloadPiece(blocks[index])

		err = b.compareHashes(index, sha1.Sum(res))

		if err != nil {
			return err
		}

		data = append(data, res...)
	}

	fmt.Println(len(data))

	err = os.WriteFile(output, data, 777)

	if err != nil {
		return err
	}

	log.Printf("Downloaded %s to %s.", b.torrent.Info.Name, output)

	return nil
}
