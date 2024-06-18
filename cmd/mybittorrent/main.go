package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Bencode struct {
	Data    string
	Size    int
	Decoded any
}

func (b Bencode) ToJson() (string, error) {

	res, err := json.Marshal(b.Decoded)

	if err != nil {
		return "", nil
	}

	return string(res), nil
}

func Decode(data string) (Bencode, error) {

	if data[0] == 'i' {
		return Int(data)
	}

	if unicode.IsDigit(rune(data[0])) {
		return String(data)
	}

	if data[0] == 'l' {
		return List(data)
	}

	return Bencode{}, errors.New("Data can't be decoded ...")

}

// i432735871e -> 432735871
func Int(data string) (Bencode, error) {

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

// - 10:hello12345 -> hello12345
func String(data string) (Bencode, error) {

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

//l5:helloi52ee -> [“hello”,52]
func List(data string) (Bencode, error) {

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

		res, err := Decode(resized[cursor:])

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

func main() {

	command := os.Args[1]

	if command == "decode" {

		bencodedValue := os.Args[2]

		bencode, err := Decode(bencodedValue)

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
