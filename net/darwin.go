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
	"syscall"
	"time"
	"unsafe"

	"github.com/sjatsh/beanstalkd-go/model"
	"github.com/sjatsh/beanstalkd-go/utils"
)

const (
	Infinity = 1 << 30
)

var kq int

func init() {
	var err error
	kq, err = syscall.Kqueue()
	if err != nil {
		panic(err)
	}
}

func sockWant(s *model.Socket, rw byte) error {
	evs := make([]syscall.Kevent_t, 0)
	ts := syscall.Timespec{}

	if s.Added {
		evs = append(evs, syscall.Kevent_t{
			Ident:  uint64(s.F.Fd()),
			Filter: s.AddedFilter,
			Flags:  syscall.EV_DELETE,
		})
	}

	if rw > 0 {
		ev := syscall.Kevent_t{
			Ident: uint64(s.F.Fd()),
		}
		switch rw {
		case 'r':
			ev.Filter = syscall.EVFILT_READ
		case 'w':
			ev.Filter = syscall.EVFILT_WRITE
		default:
			ev.Filter = syscall.EVFILT_READ
			ev.Fflags = syscall.NOTE_LOWAT
			ev.Data = Infinity
		}
		ev.Flags = syscall.EV_ADD
		ev.Udata = (*byte)(unsafe.Pointer(s))

		s.Added = true
		s.AddedFilter = ev.Filter

		evs = append(evs, ev)
	}
	if _, err := syscall.Kevent(kq, evs, nil, &ts); err != nil {
		return err
	}
	return nil
}

func sockNext(s **model.Socket, timeout time.Duration) (byte, error) {
	defer func() {
		if err := recover(); err != nil {
			utils.Log.Errorf("socket next panic error:%s\n", err)
		}
	}()

	ts := syscall.Timespec{}
	evs := make([]syscall.Kevent_t, 1)

	ts.Sec = int64(timeout / time.Second)
	ts.Nsec = int64(timeout % time.Second)

	r, err := syscall.Kevent(kq, nil, evs, &ts)
	if r == -1 || err != nil {
		return byte(r), err
	}

	ev := evs[0]
	if r > 0 {
		*s = (*model.Socket)(unsafe.Pointer(ev.Udata))
		if ev.Flags&syscall.EV_EOF != 0 {
			return 'h', nil
		}

		switch ev.Filter {
		case syscall.EVFILT_READ:
			return 'r', nil
		case syscall.EVFILT_WRITE:
			return 'w', nil
		}
	}
	return 0, nil
}
