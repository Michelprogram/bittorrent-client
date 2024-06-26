package torrent

import (
	"errors"
	"net"

	"github.com/codecrafters-io/bittorrent-starter-go/helpers"
)

type Communication struct {
	net.Conn
	PeerId string
	Ip     string
}

func (c Communication) sendInteresting() error {

	_, err := c.Write([]byte{0, 0, 0, 1, 2})

	if err != nil {
		return err
	}

	err = helpers.Wait(c, 1)

	if err != nil {
		return errors.New("not a unchoke messageback")
	}

	return nil
}

func (c Communication) Download(blocks []*Block) ([]byte, error) {

	var res []*Block

	for _, block := range blocks {
		res = append(res, block)
		err := block.Request(c.Conn)

		if err != nil {
			return nil, err
		}

	}

	return res[0].Merge(res[1:]), nil

}
