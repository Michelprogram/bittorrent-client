package command

import (
	"fmt"
	"os"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/helpers"
	"github.com/codecrafters-io/bittorrent-starter-go/internal/bencode"
	"github.com/codecrafters-io/bittorrent-starter-go/internal/torrent"
)

type Command struct{}

func (_ Command) decode() error {
	bencoded, err := bencode.NewBencode[any](os.Args[2])

	if err != nil {
		return err
	}

	json, err := bencoded.ToJson()

	if err != nil {
		return err
	}

	fmt.Println(json)

	return nil
}

func (_ Command) info() error {

	client, err := torrent.NewBittorrent(os.Args[2])

	if err != nil {
		return err
	}

	fmt.Println(client.Torrent)

	return nil
}

func (_ Command) peers() error {
	client, err := torrent.NewBittorrent(os.Args[2])

	if err != nil {
		return err
	}

	for _, ip := range client.Peers {
		fmt.Println(ip)
	}

	return nil
}

func (_ Command) handshake() error {
	client, err := torrent.NewBittorrent(os.Args[2])

	if err != nil {
		return err
	}

	communication, err := client.Handshake(os.Args[3])

	if err != nil {
		return err
	}

	fmt.Printf("Peer ID: %s\n", communication.PeerId)

	return nil
}

func (_ Command) downloadPiece() error {

	index := os.Args[5]
	path := os.Args[4]
	output := os.Args[3]

	client, err := torrent.NewBittorrent(path)

	if err != nil {
		return err
	}

	res, err := strconv.Atoi(index)

	if err != nil {
		return err
	}
	err = client.DownloadAPiece(output, res)

	if err != nil {
		return err
	}

	return nil
}

func (_ Command) download() error {
	path := os.Args[4]
	output := os.Args[3]

	client, err := torrent.NewBittorrent(path)

	if err != nil {
		return err
	}

	return helpers.ComputeTime[error](func() error {
		//return client.DownloadLow(output) //323 ms
		return client.DownloadFast(output)
	})
}

func (c Command) Run(command string) error {
	switch command {

	case "decode":
		return c.decode()
	case "info":
		return c.info()
	case "peers":
		return c.peers()
	case "handshake":
		return c.handshake()
	case "download_piece":
		return c.downloadPiece()
	case "download":
		return c.download()
	default:
		return fmt.Errorf("command not found %s\n", command)
	}
}
