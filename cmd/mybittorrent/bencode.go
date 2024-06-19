package main

import (
	"encoding/json"
	"errors"
	"unicode"
)

type Bencode struct {
	Data    string
	Size    int
	Decoded any
	Decoder
}

func (b Bencode) ToJson() (string, error) {

	res, err := json.Marshal(b.Decoded)

	if err != nil {
		return "", nil
	}

	return string(res), nil
}

func NewBencode(data string) (Bencode, error) {

	var decoder Decoder

	if data[0] == 'i' {
		decoder = DecodeInt{}
	}

	if unicode.IsDigit(rune(data[0])) {
		decoder = DecodeString{}
	}

	if data[0] == 'l' {
		decoder = DecodeList{}
	}

	if data[0] == 'd' {
		decoder = DecodeDict{}
	}

	if decoder == nil {
		return Bencode{}, errors.New("Data can't be decoded ...")
	}

	return decoder.Decode(data)
}
