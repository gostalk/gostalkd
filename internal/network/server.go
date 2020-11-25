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
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edoger/zkits-runner"

	"github.com/gostalk/gostalkd/internal/log"
)

var ErrServerClosed = errors.New("network: server closed")

func New(addr string, proc Processor) (Server, error) {
	return &tcpServer{addr: addr, proc: proc, waiter: runner.NewCloseableWaiter()}, nil
}

type Server interface {
	runner.Task
	fmt.Stringer
}

type tcpServer struct {
	mu     sync.RWMutex
	addr   string
	proc   Processor
	flag   int64
	err    error
	waiter runner.CloseableWaiter
	l      net.Listener
}

func (s *tcpServer) Execute() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch atomic.LoadInt64(&s.flag) {
	case 1:
		// The server is already running, do nothing!
		return nil
	case 2:
		return ErrServerClosed
	default:
		atomic.StoreInt64(&s.flag, 1)
	}

	socket, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go s.wait(socket)
	go s.accept(socket)

	return nil
}

func (s *tcpServer) wait(socket net.Listener) {
	defer s.waiter.Done()
	s.waiter.Wait()

	if err := socket.Close(); err != nil {
		log.Log().Warnf("network: close tcp server error %s", err)
	}
}

func (s *tcpServer) accept(socket net.Listener) {
	for {
		conn, err := socket.Accept()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				log.Log().Warnf("network: %s, retrying in 5 ms.", err)
				time.Sleep(time.Millisecond * 5)
				continue
			}
			log.Log().Errorf("network: %s", err)
			return
		}

		go s.loop(conn)
	}
}

func (s *tcpServer) loop(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	reader := bufio.NewReader(conn)
	session := newTCPSession()
	reply := newTCPReply()
	reply.conn = conn

	if err := s.proc.Connect(session, reply); err != nil {
		return
	}

	defer s.proc.Destroy(session)

	for {
		var frame []byte
		for d := time.Second; true; d += time.Second {
			if err := conn.SetReadDeadline(time.Now().Add(d)); err != nil {
				log.Log().Errorf("network: %s", err)
				return
			}

			chunk, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return
				}

				if e, ok := err.(net.Error); ok && e.Timeout() {
					chunk = trim(chunk)
					if len(chunk) > 0 {
						frame = append(frame, chunk...)
					}
					continue
				}
				return
			}

			chunk = trim(chunk)
			if len(chunk) > 0 {
				frame = append(frame, chunk...)
			}
			break
		}

		session.data = frame

		if err := s.proc.Process(session, reply); err != nil {
			if err == ErrReadContinue {
				continue
			}
			return
		}
	}
}

func (s *tcpServer) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if atomic.LoadInt64(&s.flag) != 2 {
		atomic.StoreInt64(&s.flag, 2)
		s.waiter.Close()
	}
	return nil
}

func (s *tcpServer) String() string {
	return s.addr
}

func trim(b []byte) []byte {
	if n := len(b); n > 1 && b[n-2] == '\r' && b[n-1] == '\n' {
		return b[:n-2]
	}
	return b
}
