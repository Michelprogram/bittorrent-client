package main

import (
	"encoding/json"
	"fmt"
)

type Bencode[T any] struct {
	Data    string
	Size    int
	Decoded T
}

func (b Bencode[T]) ToJson() (string, error) {

	res, err := json.Marshal(b.Decoded)

	if err != nil {
		return "", nil
	}

	return string(res), nil
}

func NewBencode[T any](data string) (Bencode[T], error) {

	if data == "" {
		return Bencode[T]{}, fmt.Errorf("data is empty")
	}

	var bencode Bencode[T]

	switch data[0] {
	case 'i':
		var decoder DecodeInt
		decode, err := decoder.Decode(data)

		if err != nil {
			return Bencode[T]{}, err
		}

		bencode.Decoded = any(decode.Decoded).(T)
		bencode.Data = decode.Data
		bencode.Size = decode.Size

	case 'l':
		var decoder DecodeList
		decode, err := decoder.Decode(data)

		if err != nil {
			return Bencode[T]{}, err
		}

		bencode.Decoded = any(decode.Decoded).(T)
		bencode.Data = decode.Data
		bencode.Size = decode.Size
	case 'd':
		var decoder DecodeDict
		decode, err := decoder.Decode(data)

		if err != nil {
			return Bencode[T]{}, err
		}

		bencode.Decoded = any(decode.Decoded).(T)
		bencode.Data = decode.Data
		bencode.Size = decode.Size
	default:
		var decoder DecodeString
		decode, err := decoder.Decode(data)

		if err != nil {
			return Bencode[T]{}, err
		}

		bencode.Decoded = any(decode.Decoded).(T)
		bencode.Data = decode.Data
		bencode.Size = decode.Size
	}

	return bencode, nil

}
