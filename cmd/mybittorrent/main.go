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

		fmt.Printf("Tracker URL: %s\nLength: %d\nInfo Hash: %x\n", torrent.Announce, torrent.Info.Length, torrent.Hash)

	} else {
		panic("Unknown command: " + command)
	}
}
