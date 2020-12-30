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
	err error
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

// Err
func (t *Tokens) Err() error {
	return t.err
}

// Len
func (t *Tokens) Len() int {
	return len(t.t)
}

// NextString
func (t *Tokens) NextString(s *string) *Tokens {
	if t.err != nil {
		return t
	}
	if t.outOfBounds() {
		t.err = IndexOutOfBounds
		return t
	}
	r := t.t[t.cur]
	t.cur++
	*s = r
	return t
}

// NextInt64
func (t *Tokens) NextInt64(i *int64) *Tokens {
	if t.err != nil {
		return t
	}
	if t.outOfBounds() {
		t.err = IndexOutOfBounds
		return t
	}
	r := t.t[t.cur]
	t.cur++
	ri, err := strconv.ParseInt(r, 10, 64)
	t.err = err
	*i = ri
	return t
}

// NextInt
func (t *Tokens) NextInt(i *int) *Tokens {
	if t.err != nil {
		return t
	}
	var r int64
	t.NextInt64(&r)
	*i = int(r)
	return t
}

// NextUint64
func (t *Tokens) NextUint64(i *uint64) *Tokens {
	if t.err != nil {
		return t
	}
	var r int64
	t.NextInt64(&r)
	*i = uint64(r)
	return t
}

// NextUint32
func (t *Tokens) NextUint32(i *uint32) *Tokens {
	if t.err != nil {
		return t
	}
	var r int64
	t.NextInt64(&r)
	*i = uint32(r)
	return t
}

// NextFloat64
func (t *Tokens) NextFloat64(f *float64) *Tokens {
	if t.err != nil {
		return t
	}
	if t.outOfBounds() {
		t.err = IndexOutOfBounds
		return t
	}
	r := t.t[t.cur]
	t.cur++
	of, err := strconv.ParseFloat(r, 64)
	*f = of
	t.err = err
	return t
}

// NextFloat32
func (t *Tokens) NextFloat32(f *float32) *Tokens {
	if t.err != nil {
		return t
	}
	var r float64
	t.NextFloat64(&r)
	*f = float32(r)
	return t
}

func (t *Tokens) outOfBounds() bool {
	return t.cur >= len(t.t)
}
