// Copyright 2020 The Gostalkd Project Authors.
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

package str

import (
	"errors"
	"strconv"
	"strings"
)

var IndexOutOfBounds = errors.New("index out of bounds")

type Tokens struct {
	cur int
	t   []string
}

// Parse
func Parse(cmd string) *Tokens {
	tokens := strings.Split(cmd, " ")
	return &Tokens{
		t: tokens,
	}
}

// Len
func (t *Tokens) Len() int {
	return len(t.t)
}

// NextString
func (t *Tokens) NextString() (string, error) {
	if t.outOfBounds() {
		return "", IndexOutOfBounds
	}
	r := t.t[t.cur]
	t.cur++
	return r, nil
}

// NextInt64
func (t *Tokens) NextInt64() (int64, error) {
	if t.outOfBounds() {
		return 0, IndexOutOfBounds
	}
	r := t.t[t.cur]
	t.cur++
	return strconv.ParseInt(r, 10, 64)
}

// NextInt
func (t *Tokens) NextInt() (int, error) {
	r, err := t.NextInt64()
	return int(r), err
}

// NextUint64
func (t *Tokens) NextUint64() (uint64, error) {
	r, err := t.NextInt64()
	return uint64(r), err
}

// NextUint32
func (t *Tokens) NextUint32() (uint32, error) {
	r, err := t.NextInt64()
	return uint32(r), err
}

// NextFloat64
func (t *Tokens) NextFloat64() (float64, error) {
	if t.outOfBounds() {
		return 0, IndexOutOfBounds
	}
	r := t.t[t.cur]
	t.cur++
	return strconv.ParseFloat(r, 64)
}

// NextFloat32
func (t *Tokens) NextFloat32() (float32, error) {
	r, err := t.NextFloat64()
	return float32(r), err
}

func (t *Tokens) outOfBounds() bool {
	return t.cur >= len(t.t)
}
