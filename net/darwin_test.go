// +build darwin

// Copyright 2020 SunJun <i@sjis.me>
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
package net

import (
	"bytes"
	"io"
	net2 "net"
	"testing"

	"github.com/gostalk/gostalkd/model"
)

func TestSockWant(t *testing.T) {
	addr := "127.0.0.1:9999"
	a, err := net2.ResolveTCPAddr("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net2.ListenTCP("tcp", a)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	f, err := l.File()
	if err != nil {
		t.Fatal(err)
	}

	if err := sockWant(&model.Socket{
		F:  f,
		Ln: l,
		H: func(i interface{}, i2 byte) {
			if i == nil {
				return
			}
			c, err := i.(*model.Socket).Ln.Accept()
			if err != nil {
				t.Fatal(err)
			}
			t.Log("get client request")

			b := make([]byte, 1024)
			res := bytes.Buffer{}
			for {
				size, err := c.Read(b)
				if err == io.EOF {
					break
				}
				res.Write(b[:size])
			}
			t.Log(string(b))
		},
	}, 'r'); err != nil {
		t.Fatal(err)
	}

	// 模拟客户端写入
	c, err := net2.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	c.Close()

	// 获取客户端socket
	var sock *model.Socket
	rw, err := sockNext(&sock, 1)
	if err != nil {
		t.Fatal(err)
	}
	sock.H(sock, rw)
}
