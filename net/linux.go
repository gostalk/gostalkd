// +build linux

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
	"syscall"
	"time"
	"unsafe"

	"github.com/sjatsh/beanstalkd-go/model"
)

var epfd int

func init() {
	var err error
	epfd, err = syscall.EpollCreate(1)
	if err != nil {
		panic(err)
	}
}

func sockWant(s *model.Socket, rw byte) error {
	var op int
	if !s.Added && rw <= 0 {
		return nil
	}
	if !s.Added && rw > 0 {
		s.Added = true
		op = syscall.EPOLL_CTL_ADD
	} else if rw <= 0 {
		op = syscall.EPOLL_CTL_DEL
	} else {
		op = syscall.EPOLL_CTL_MOD
	}

	ev := syscall.EpollEvent{}
	switch rw {
	case 'r':
		ev.Events = syscall.EPOLLIN
	case 'w':
		ev.Events = syscall.EPOLLOUT
	}
	ev.Events |= syscall.EPOLLRDHUP | syscall.EPOLLPRI
	ev.Pad = *(*int32)(unsafe.Pointer(s))
	return syscall.EpollCtl(epfd, op, int(s.F.Fd()), &ev)
}

func sockNext(s **model.Socket, timeout time.Duration) (byte, error) {
	ev := syscall.EpollEvent{}
	r, err := syscall.EpollWait(epfd, []syscall.EpollEvent{ev}, (int)(timeout.Seconds()))
	if r == -1 || err != nil {
		return byte(r), err
	}
	if r > 0 {
		*s = (*model.Socket)(unsafe.Pointer(&ev.Pad))
		if ev.Events&(syscall.EPOLLHUP|syscall.EPOLLRDHUP) > 0 {
			return 'h', nil
		}
		if ev.Events&syscall.EPOLLIN > 0 {
			return 'r', nil
		}
		if ev.Events&syscall.EPOLLOUT > 0 {
			return 'w', nil
		}
	}
	return 0, nil
}
