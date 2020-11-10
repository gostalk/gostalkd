// Package json provides json operations wrapping ignoring stdlib or third-party lib json.
package json

import (
	"encoding/json"
	"io"
)

// ConfigCompatibleWithStandardLibrary tries to be 50% compatible
// with standard library behavior.
// 50% - -!
//var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal adapts to json/encoding Marshal API.
//
// Marshal returns the JSON encoding of v, adapts to json/encoding Marshal API
// Refer to https://godoc.org/encoding/json#Marshal for more information.
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalToString convenient method to write as string instead of []byte.
func MarshalToString(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	return string(bytes), err
}

// MarshalIndent same as json.MarshalIndent. Prefix is not supported.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// UnmarshalFromString is a convenient method to read from string instead of []byte.
func UnmarshalFromString(str string, v interface{}) error {
	return json.Unmarshal([]byte(str), v)
}

// Unmarshal adapts to json/encoding Unmarshal API
//
// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
// Refer to https://godoc.org/encoding/json#Unmarshal for more information.
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewEncoder same as json.NewEncoder
func NewEncoder(writer io.Writer) *json.Encoder {
	return json.NewEncoder(writer)
}

// NewDecoder adapts to json/stream NewDecoder API.
//
// NewDecoder returns a new decoder that reads from r.
//
// Instead of a json/encoding Decoder, an Decoder is returned
// Refer to https://godoc.org/encoding/json#NewDecoder for more information.
func NewDecoder(reader io.Reader) *json.Decoder {
	return json.NewDecoder(reader)
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return json.Valid(data)
}
