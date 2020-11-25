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

package main

import (
	"fmt"

	"github.com/edoger/zkits-runner"

	"github.com/gostalk/gostalkd/internal/network"
)

func main() {
	fmt.Println("Hello, Gostalkd!")

	s, _ := network.New("127.0.0.1:5056", new(processor))
	runner.New().MustRun(s).Wait()
}

type processor struct {
}

func (*processor) Connect(s network.Session, r network.Reply) error {
	fmt.Println(s.ID(), "Connected~")
	_, _ = r.WriteString("Welcome!\r\n")
	if err := r.Flush(); err != nil {
		fmt.Println("Reply.Flush", err)
	}
	return nil
}

func (*processor) Process(s network.Session, r network.Reply) error {
	fmt.Println(s.ID(), string(s.Data()))
	_, _ = r.WriteString("Ok~~~\r\n")
	if err := r.Flush(); err != nil {
		fmt.Println("Reply.Flush", err)
	}
	return network.ErrReadContinue
}

func (*processor) Destroy(s network.Session) {
	fmt.Println(s.ID(), "Bye~")
}
