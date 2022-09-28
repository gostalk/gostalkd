// Copyright 2020 gostalkd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"errors"
	"strconv"
)

func StrLen(s []byte) int {
	l := bytes.Index(s, []byte("\r"))
	if l == -1 {
		l = len(s)
	}
	return l
}

func ToStr(s []byte) string {
	idx := bytes.Index(s, []byte("\r"))
	return string(s[:idx])
}

func StrTol(str []byte) (int64, error) {
	data := make([]byte, 0)
	for _, b := range str {
		if b < '0' || b > '9' {
			break
		}
		data = append(data, b)
	}
	return strconv.ParseInt(string(data), 10, 64)
}

func ReadInt(buf []byte, idx *int) (int64, error) {
	begin := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] != ' ' {
			break
		}
		begin++
	}
	if buf[begin] < '0' || buf[begin] > '9' {
		return 0, errors.New("fmt error")
	}

	var data []byte
	end := begin
	for ; end < len(buf); end++ {
		if buf[end] == ' ' || buf[end] < '0' || buf[end] > '9' {
			break
		}
		data = append(data, buf[end])
	}
	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	*idx += end
	return i, nil
}

func ReadDuration(buf []byte, idx *int) (int64, error) {
	t, err := ReadInt(buf, idx)
	if err != nil {
		return 0, err
	}
	duration := t * 1000000000
	return duration, nil
}

func ReadTubeName(buf []byte, idx *int) ([]byte, error) {
	begin := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] != ' ' {
			break
		}
		begin++
	}

	l := StrSpn(buf[begin:])
	if l == 0 {
		return nil, errors.New("is invalid tube name")
	}
	*idx += begin + l
	return buf[begin : l+begin], nil
}
