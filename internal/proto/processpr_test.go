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

package proto_test

import (
	"fmt"
	"testing"

	"github.com/gostalk/gostalkd/internal/proto"
)

type testSession struct {
	id   uint64
	data []byte
}

func (s *testSession) ID() uint64 {
	return s.id
}

func (s *testSession) Data() []byte {
	return s.data
}

func (s *testSession) String() string {
	return fmt.Sprintf("id:%d", s.id)
}

type testReply struct {
	t *testing.T
}

func (r *testReply) WriteByte(c byte) error {
	r.t.Log(c)
	return nil
}

func (r *testReply) WriteString(s string) (int, error) {
	r.t.Log(s)
	return len(s), nil
}

func (r *testReply) Write(bytes []byte) (int, error) {
	r.t.Log(string(bytes))
	return len(bytes), nil
}

func (r *testReply) Flush() error {
	r.t.Log("flushed")
	return nil
}

func TestNewCmdProcessor(t *testing.T) {
	proto.NewCmdProcessor()
}

func TestCmdProcessor_Connect(t *testing.T) {
	p := proto.NewCmdProcessor()

	session := &testSession{
		id:   1,
		data: []byte("hello"),
	}
	reply := &testReply{
		t: t,
	}
	p.Connect(session, reply)
}

func TestCmdProcessor_Process(t *testing.T) {
	p := proto.NewCmdProcessor()

	session := &testSession{
		id:   1,
		data: []byte("put 1 10 10 5"),
	}
	reply := &testReply{
		t: t,
	}
	p.Connect(session, reply)
	p.Process(session, reply)

	session.data = []byte("hello")
	p.Process(session, reply)
}

func TestCmdProcessor_Destroy(t *testing.T) {
	p := proto.NewCmdProcessor()

	session := &testSession{
		id:   1,
		data: []byte("put 1 10 10 5"),
	}
	reply := &testReply{
		t: t,
	}
	p.Connect(session, reply)
	p.Process(session, reply)

	session.data = []byte("hello")
	p.Process(session, reply)
	p.Destroy(session)
}
