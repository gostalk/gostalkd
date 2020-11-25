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
	"fmt"
	"sync/atomic"
)

// This is the global connection session counter.
// For each new network connection, a globally unique id will be assigned.
var sid uint64

type Session interface {
	fmt.Stringer

	ID() uint64
	Data() []byte
}

func newTCPSession() *tcpSession {
	return &tcpSession{id: atomic.AddUint64(&sid, 1)}
}

type tcpSession struct {
	id   uint64
	data []byte
}

func (s *tcpSession) ID() uint64 {
	return s.id
}

func (s *tcpSession) Data() []byte {
	return s.data
}

func (s *tcpSession) String() string {
	return fmt.Sprintf("network:session:%d", s.id)
}
