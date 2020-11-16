// +build windows

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
package core

import (
	"os"

	"golang.org/x/sys/windows"

	"github.com/sjatsh/beanstalkd-go/model"
	"github.com/sjatsh/beanstalkd-go/utils"
)

// WalDirLock
func WalDirLock(w *model.Wal) bool {
	if err := os.MkdirAll(w.Dir, os.ModePerm); err != nil {
		utils.Log.Errorf("mkdir err: %s", err)
		return false
	}
	path := w.Dir + "/lock"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		utils.Log.Warnf("open %s err:%s", path, err)
		return false
	}
	err = windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &windows.Overlapped{})
	if err == nil {
		return true
	}
	return false
}
