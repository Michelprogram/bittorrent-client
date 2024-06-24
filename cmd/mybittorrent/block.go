package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

type Block struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
	Data         []byte
}

type blockWithoutData struct {
	lengthPrefix uint32
	id           uint8
	index        uint32
	begin        uint32
	length       uint32
}

func (b *Block) Request(conn net.Conn) error {

	var buffer bytes.Buffer

	bwt := blockWithoutData{
		lengthPrefix: b.lengthPrefix,
		id:           b.id,
		index:        b.index,
		begin:        b.begin,
		length:       b.length,
	}

	binary.Write(&buffer, binary.BigEndian, bwt)

	_, err := conn.Write(buffer.Bytes())

	if err != nil {
		return err
	}

	reader := make([]byte, 4)

	_, err = conn.Read(reader)

	size := binary.BigEndian.Uint32(reader)

	reader = make([]byte, size)

	_, err = io.ReadFull(conn, reader)

	if err != nil {
		return err
	}

	b.Data = reader[9:]

	return nil

}

func (b Block) Merge(blocks []*Block) []byte {
	merged := b.Data

	for _, block := range blocks {
		merged = append(merged, block.Data...)
	}

	return merged

}
