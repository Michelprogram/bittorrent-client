package main

import (
	"os"
)

type Bittorrent struct {
}

type Info struct {
	Length      int
	Name        string
	PieceLength int
	Pieces      string
}

type Torrent struct {
	Announce  string
	CreatedBy string
	Info
}

func NewBittorrent() *Bittorrent {
	return &Bittorrent{}
}

func (b Bittorrent) Info(path string) (Torrent, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return Torrent{}, nil
	}

	info, err := NewBencode[map[string]any](string(data))

	if err != nil {
		return Torrent{}, nil
	}
	return Torrent{
		Announce:  info.Decoded["announce"].(string),
		CreatedBy: info.Decoded["created by"].(string),
		Info: Info{
			Length:      info.Decoded["info"].(map[string]any)["length"].(int),
			Name:        info.Decoded["info"].(map[string]any)["name"].(string),
			PieceLength: info.Decoded["info"].(map[string]any)["piece length"].(int),
			Pieces:      info.Decoded["info"].(map[string]any)["pieces"].(string),
		},
	}, nil
}

func (b Bittorrent) Receive(data string) (Bencode[any], error) {

	return NewBencode[any](data)
}
