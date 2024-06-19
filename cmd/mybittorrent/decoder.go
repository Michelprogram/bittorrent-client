package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Decoder interface {
	Decode(data string) (Bencode, error)
}

type DecodeList struct{}

//l5:helloi52ee -> [“hello”,52]
func (_ DecodeList) Decode(data string) (Bencode, error) {

	if data == "le" {
		return Bencode{
			Data:    data,
			Decoded: make([]interface{}, 0),
			Size:    0,
		}, nil
	}

	decoded := make([]interface{}, 0)

	resized := data[1 : len(data)-1]

	max := len(data) - 2
	cursor := 0
	flag := true

	for flag {

		res, err := NewBencode(resized[cursor:])

		if err != nil {
			return Bencode{}, err
		}

		cursor += res.Size

		decoded = append(decoded, res.Decoded)

		if cursor >= max || resized[cursor] == 'e' {
			flag = false
		}
	}

	return Bencode{
		Data:    data,
		Decoded: decoded,
		Size:    cursor + 2,
	}, nil

}

type DecodeInt struct{}

// i432735871e -> 432735871
func (_ DecodeInt) Decode(data string) (Bencode, error) {

	end := strings.Index(data, "e")
	res, err := strconv.Atoi(data[1:end])

	if err != nil {
		return Bencode{}, nil
	}

	return Bencode{
		Data:    data,
		Decoded: res,
		Size:    len(fmt.Sprintf("%d", res)) + 2,
	}, nil

}

type DecodeString struct{}

// - 10:hello12345 -> hello12345
func (_ DecodeString) Decode(data string) (Bencode, error) {

	end := strings.Index(data, ":")

	length, err := strconv.Atoi(data[:end])
	if err != nil {
		return Bencode{}, err
	}

	return Bencode{
		Data:    data,
		Decoded: data[end+1 : end+length+1],
		Size:    length + end + 1,
	}, nil

}

type DecodeDict struct{}

func (d DecodeDict) Decode(data string) (Bencode, error) {

	if data == "de" {
		return Bencode{
			Data:    data,
			Decoded: make(map[string]any, 0),
			Size:    0,
		}, nil
	}

	decoded := make(map[string]any, 0)

	max := len(data) - 2
	cursor := 1
	flag := true

	for flag {

		key, err := NewBencode(data[cursor:])

		if err != nil {
			return Bencode{}, err
		}

		cursor += key.Size

		value, err := NewBencode(data[cursor:])
		if err != nil {
			return Bencode{}, err
		}

		cursor += value.Size

		decoded[key.Decoded.(string)] = value.Decoded

		if cursor >= max || data[cursor] == 'e' {
			flag = false
		}
	}

	return Bencode{
		Data:    data,
		Decoded: decoded,
		Size:    cursor + 2,
	}, nil

}
