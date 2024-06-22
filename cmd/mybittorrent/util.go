package main

import "net"

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
