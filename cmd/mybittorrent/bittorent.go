package main

import (
	"crypto/sha1"
	"os"

	"github.com/jackpal/bencode-go"
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
