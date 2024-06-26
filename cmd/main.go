package main

import (
	"log"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/internal/command"
)

func main() {

	var command command.Command

	err := command.Run(os.Args[1])

	if err != nil {
		log.Fatal(err)
	}

}
