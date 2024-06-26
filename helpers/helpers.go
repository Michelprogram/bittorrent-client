package helpers

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	"math/rand"
)

var (
	SIZE_PEER_ID = 20
)

func Wait(conn net.Conn, id uint8) error {

	response := make([]byte, 5)

	for response[4] != byte(id) {
		_, err := conn.Read(response)
		if err != nil {
			return err
		}
	}

	return nil
}

func RandomPeerId() string {

	var buffer bytes.Buffer

	numbers := [10]byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	}

	for i := 0; i < SIZE_PEER_ID; i++ {
		buffer.WriteByte(numbers[rand.Intn(len(numbers))])
	}

	return buffer.String()

}

func CompareHashes(index int, input, hashes [20]byte) error {

	if !bytes.Equal(input[:], hashes[:]) {
		return fmt.Errorf("hash doesn't match at index %d : \nPiece hash : %x\nDownloaded hash :%x\n", index, input, hashes)
	}

	return nil

}

func ComputeTime[T any](cb func() T) T {
	start := time.Now()

	res := cb()

	log.Println(time.Since(start))

	return res
}
