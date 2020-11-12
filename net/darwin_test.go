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
	"fmt"
	"io"
	net2 "net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sjatsh/beanstalk-go/model"
)

func TestPutAndReserveCmd(t *testing.T) {
	// 模拟客户端写入
	c, err := net2.Dial("tcp", "127.0.0.1:11400")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// put 一条3秒延时的job
	if _, err := c.Write([]byte("put 1 3 10 5\r\nhello\r\n")); err != nil {
		t.Fatal(err)
	}
	// 并且同时去reserve
	if _, err := c.Write([]byte("reserve\r\n")); err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		reply := make([]byte, 1024)
		for {
			r, err := c.Read(reply)
			if err != nil {
				t.Fatal(err)
			}
			if r == 0 {
				continue
			}
			reply = reply[:r]
			fmt.Print(string(reply))
			if strings.Contains(string(reply), "hello") {
				break
			}
		}
	}()
	wg.Wait()
}

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
	f, err := l.File()
	if err != nil {
		t.Fatal(err)
	}
	l.Close()

	if err := sockWant(&model.Socket{
		F: f,
		H: func(i interface{}, i2 byte) {
			s := i.(*model.Socket)
			l, err := net2.FileListener(s.F)
			if err != nil {
				t.Fatal(err)
			}
			c, err := l.Accept()
			if err != nil {
				t.Fatal(err)
			}

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
			t.Log("get client request")
			t.Logf("%#v", i)
			t.Logf("%s", string(i2))
		},
	}, 'r'); err != nil {
		t.Fatal(err)
	}

	// 模拟客户端写入
	c, err := net2.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Write([]byte("hello word")); err != nil {
		t.Fatal(err)
	}
	c.Close()

	// 获取客户端socket
	sock := &model.Socket{}
	rw, err := sockNext(&sock, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	sock.H(sock, rw)
}
