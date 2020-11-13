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
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/utils"
)

var client net.Conn

func TestMain(m *testing.M) {

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	s, err := NewServer(
		model.WithPort(*utils.Port),
		model.WithAddr(*utils.ListenAddr),
		model.WithUser(*utils.User),
	)
	if err != nil {
		panic(err)
	}
	go Start(s)

	time.Sleep(time.Second)

	client, err = net.Dial("tcp", "127.0.0.1:11400")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go func() {
		reply := make([]byte, 1024)
		for {
			r, err := client.Read(reply)
			if err != nil && err != io.EOF {
				return
			}
			if r == 0 {
				continue
			}
			reply = reply[:r]
			fmt.Print(string(reply))
		}
	}()
	m.Run()
}
