package main

import (
	"reflect"
	"testing"
)

type Tests struct {
	name  string
	input string
	want  any
}

// Test decode int
func TestDecodeInt(t *testing.T) {

	// Defining the columns of the table
	var tests = []Tests{
		// the table itself
		{"Test 1", "i52e", 52},
		{"Test 2", "i-52e", -52},
		{"Test 3", "i0e", 0},
		{"Test 4", "i-0e", 0},
		{"Test 5", "i456787e", 456787},
		{"Test 6", "i-456787e", -456787},
	}
	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := DecodeInt{}
			res, err := decoder.Decode(tt.input)

			if err != nil {
				t.Error(err)
			}

			if res.Decoded != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, res.Decoded)
			}
		})
	}

}

// Test decode list
func TestDecodeList(t *testing.T) {

	// Defining the columns of the table
	var tests = []Tests{
		{"Test 1", "l5:helloi52ee", []any{
			"hello", 52,
		}},
		{"Test 2", "l5:helloi52e5:worldi-52ee", []any{
			"hello", 52, "world", -52,
		}},
		{"Test 3", "le", []any{}},
		{"Test 4", "ll5:helloei5ee", []any{
			[]any{"hello"}, 5,
		}},
	}

	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := DecodeList{}
			res, err := decoder.Decode(tt.input)

			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(res.Decoded, tt.want) {
				t.Errorf("Expected %d, got %d", tt.want, res.Decoded)
			}
		})
	}
}

// Test decode string

func TestDecodeString(t *testing.T) {

	// Defining the columns of the table
	var tests = []Tests{
		// the table itself
		{"Test 1", "5:hello", "hello"},
		{"Test 2", "5:world", "world"},
		{"Test 3", "0:", ""},
	}
	// The execution loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := DecodeString{}
			res, err := decoder.Decode(tt.input)

			if err != nil {
				t.Error(err)
			}

			if res.Decoded != tt.want {
				t.Errorf("Expected %d, got %s", tt.want, res.Decoded)
			}
		})
	}
}

// Test decode dict
func TestDecodeDict(t *testing.T) {

	var tests = []Tests{
		{"Test 1", "d3:foo3:bar5:helloi52ee", map[string]any{
			"foo":   "bar",
			"hello": 52,
		}},
		{"Test 2", "de", map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := DecodeDict{}
			res, err := decoder.Decode(tt.input)

			if err != nil {
				t.Error(err)
			}

			for key, value := range res.Decoded {
				if tt.want.(map[string]interface{})[key] != value {
					t.Errorf("Expected %v, got %v", tt.want.(map[string]interface{})[key], value)
				}
			}
		})
	}

}
