package main

import (
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

var (
	HASH_LEN = 20
)

type Bittorrent struct {
}

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

	for _, hash := range t.piecesHash() {
		hashesString += fmt.Sprintf("%x\n", hash)
	}

	return fmt.Sprintf("Tracker URL: %s\nLength: %d\nInfo Hash: %x\nPiece Length: %d\nPiece Hashes:\n%s", t.Announce, t.Info.Length, t.Hash, t.Info.PieceLength, hashesString)
}

func NewBittorrent() *Bittorrent {
	return &Bittorrent{}
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
