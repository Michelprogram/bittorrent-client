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

	} else {
		panic("Unknown command: " + command)
	}
}
