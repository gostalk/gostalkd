//go:build windows
// +build windows

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
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
)

var sysInfo *SysInfo

func init() {
	info, err := host.Info()
	if err != nil {
		Log.Panic(err)
	}
	sysInfo = &SysInfo{}
	sysInfo.SysName = info.OS
	sysInfo.Machine = info.HostID
	sysInfo.NodeName = info.Hostname
	sysInfo.Version = info.PlatformVersion
}

// GetSysInfo
func GetSysInfo() *SysInfo {
	return &SysInfo{
		SysName:  sysInfo.SysName,
		Machine:  sysInfo.Machine,
		NodeName: sysInfo.NodeName,
		Version:  sysInfo.Version,
		Release:  sysInfo.Release,
	}
}

// Getrusage
func Getrusage() *Rusage {
	ru := &Rusage{}
	rets, err := cpu.Times(false)
	if err != nil || len(rets) <= 0 {
		return ru
	}
	ret := rets[0]
	ru.Usec = int64(ret.User)
	ru.Ssec = int64(ret.System)
	ru.Total = int64(ret.Total())
	return ru
}
