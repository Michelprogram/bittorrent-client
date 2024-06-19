package main

type Bittorrent struct {
}

func NewBittorrent() *Bittorrent {
	return &Bittorrent{}
}

func (b Bittorrent) Receive(data string) (Bencode, error) {

	return NewBencode(data)
}
