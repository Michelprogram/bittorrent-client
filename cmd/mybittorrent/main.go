package main

import (
	"fmt"
	"log"
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

		log.Println(torrent)

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

	} else {
		panic("Unknown command: " + command)
	}
}
