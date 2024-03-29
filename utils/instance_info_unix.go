//go:build linux || darwin || openbsd

// Copyright 2020 gostalkd
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

package utils

import (
	"os"

	"golang.org/x/sys/unix"
)

var systemInfo *SysInfo

func init() {
	uname := &unix.Utsname{}
	if err := unix.Uname(uname); err != nil {
		Log.Panic(err)
	}
	systemInfo = &SysInfo{
		SysName:  string(uname.Sysname[:]),
		NodeName: string(uname.Nodename[:]),
		Release:  string(uname.Release[:]),
		Version:  string(uname.Version[:]),
		Machine:  string(uname.Version[:]),
	}
}

// GetSysInfo
func GetSysInfo() *SysInfo {
	return &SysInfo{
		SysName:  systemInfo.SysName,
		NodeName: systemInfo.NodeName,
		Release:  systemInfo.Release,
		Version:  systemInfo.Version,
		Machine:  systemInfo.Machine,
	}
}

// Getrusage
func Getrusage() *Rusage {
	ru := &unix.Rusage{}
	_ = unix.Getrusage(os.Getpid(), ru)
	return &Rusage{
		Usec:  ru.Utime.Sec,
		Ssec:  ru.Stime.Sec,
		Total: int64(ru.Utime.Usec),
	}
}
