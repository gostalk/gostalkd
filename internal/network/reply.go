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

package network

import (
	"bytes"
	"io"
	"net"
	"sync"
)

type Reply interface {
	io.ByteWriter
	io.Writer
	io.StringWriter

	Flush() error
}

type tcpReplyPool struct {
	c chan *tcpReply
}

func (p *tcpReplyPool) get() *tcpReply {
	select {
	case r := <-p.c:
		return r
	default:
		return new(tcpReply)
	}
}

func (p *tcpReplyPool) put(r *tcpReply) {
	if r.buf.Cap() >= 4096 {
		r.buf = bytes.Buffer{}
	} else {
		r.buf.Reset()
	}
	r.conn = nil
	select {
	case p.c <- r:
		return
	default:
	}
}

func newTCPReply() *tcpReply {
	return &tcpReply{}
}

type tcpReply struct {
	buf   bytes.Buffer
	conn  net.Conn
	mutex sync.Mutex
}

func (r *tcpReply) WriteByte(c byte) error {
	return r.buf.WriteByte(c)
}

func (r *tcpReply) Write(p []byte) (int, error) {
	return r.buf.Write(p)
}

func (r *tcpReply) WriteString(s string) (int, error) {
	return r.buf.WriteString(s)
}

func (r *tcpReply) Flush() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.buf.Len() == 0 {
		return nil
	}

	defer r.buf.Reset()
	data := r.buf.Bytes()
	n, err := r.conn.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}
