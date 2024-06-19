package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Decoder[T any] interface {
	Decode(data string) (Bencode[T], error)
}

type DecodeList struct{}

//l5:helloi52ee -> [“hello”,52]
func (d DecodeList) Decode(data string) (Bencode[[]interface{}], error) {

	if data == "le" {
		return Bencode[[]interface{}]{
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

		res, err := NewBencode[any](resized[cursor:])

		if err != nil {
			return Bencode[[]interface{}]{}, err
		}

		cursor += res.Size

		decoded = append(decoded, res.Decoded)

		if cursor >= max || resized[cursor] == 'e' {
			flag = false
		}
	}

	return Bencode[[]interface{}]{
		Data:    data,
		Decoded: decoded,
		Size:    cursor + 2,
	}, nil

}

type DecodeInt struct{}

// i432735871e -> 432735871
func (_ DecodeInt) Decode(data string) (Bencode[int], error) {

	end := strings.Index(data, "e")
	res, err := strconv.Atoi(data[1:end])

	if err != nil {
		return Bencode[int]{}, nil
	}

	return Bencode[int]{
		Data:    data,
		Decoded: res,
		Size:    len(fmt.Sprintf("%d", res)) + 2,
	}, nil

}

type DecodeString struct{}

// - 10:hello12345 -> hello12345
func (_ DecodeString) Decode(data string) (Bencode[string], error) {

	end := strings.Index(data, ":")

	length, err := strconv.Atoi(data[:end])
	if err != nil {
		return Bencode[string]{}, err
	}

	return Bencode[string]{
		Data:    data,
		Decoded: data[end+1 : end+length+1],
		Size:    length + end + 1,
	}, nil

}

type DecodeDict struct{}

func (d DecodeDict) Decode(data string) (Bencode[map[string]any], error) {

	if data == "de" {
		return Bencode[map[string]any]{
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

		key, err := NewBencode[any](data[cursor:])

		if err != nil {
			return Bencode[map[string]any]{}, err
		}

		cursor += key.Size

		value, err := NewBencode[any](data[cursor:])
		if err != nil {
			return Bencode[map[string]any]{}, err
		}

		cursor += value.Size

		decoded[key.Decoded.(string)] = value.Decoded

		if cursor >= max || data[cursor] == 'e' {
			flag = false
		}
	}

	return Bencode[map[string]any]{
		Data:    data,
		Decoded: decoded,
		Size:    cursor + 2,
	}, nil

}
