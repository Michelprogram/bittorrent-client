package main

import (
	"fmt"
	"os"
)

func main() {

	bittorent := NewBittorrent()

	command := os.Args[1]

	if command == "decode" {

		bencodedValue := os.Args[2]

		bencode, err := bittorent.Receive(bencodedValue)

		if err != nil {
			panic(err)
		}

		json, err := bencode.ToJson()
		if err != nil {
			panic(err)
		}

		fmt.Println(json)

	} else if command == "info" {

		path := os.Args[2]
		torrent, err := bittorent.Info(path)

		if err != nil {
			panic(err)
		}

		fmt.Println(torrent)

	} else if command == "peers" {

		path := os.Args[2]
		torrent, err := bittorent.Info(path)

		if err != nil {
			panic(err)
		}

		tracker, err := NewTracker(torrent)
		if err != nil {
			panic(err)
		}

		response, err := tracker.Get()
		if err != nil {
			panic(err)
		}

		for _, ip := range response.Peers {
			fmt.Println(ip)
		}

	} else if command == "handshake" {
		path := os.Args[2]
		peer := os.Args[3]

		torrent, err := bittorent.Info(path)

		if err != nil {
			panic(err)
		}

		err = bittorent.Handshake(peer, torrent)

		if err != nil {
			panic(err)
		}

	} else if command == "download_piece" {
		path := os.Args[4]
		output := os.Args[3]

		torrent, err := bittorent.Info(path)

		if err != nil {
			panic(err)
		}

		tracker, err := NewTracker(torrent)
		if err != nil {
			panic(err)
		}

		response, err := tracker.Get()
		if err != nil {
			panic(err)
		}

		err = bittorent.Handshake(response.Peers[1].String(), torrent)

		if err != nil {
			panic(err)
		}

		err = bittorent.Download(output)

		if err != nil {
			panic(err)
		}

	} else {
		panic("Unknown command: " + command)
	}
}
