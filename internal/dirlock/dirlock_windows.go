// +build windows

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

package dirlock

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func (l *DirLock) Lock() error {
	f, err := os.Open(l.dir)
	if err != nil {
		return err
	}
	l.f = f
	if err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &windows.Overlapped{}); err != nil {
		return fmt.Errorf("cannot flock directory %s - %s (possibly in use by another instance)", l.dir, err)
	}
	return nil
}

func (l *DirLock) Unlock() error {
	defer l.f.Close()
	return windows.UnlockFileEx(windows.Handle(l.f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 1, 0, &windows.Overlapped{})
}
