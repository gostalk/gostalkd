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
	"fmt"
	"math/rand"

	"golang.org/x/sys/unix"

	"github.com/sjatsh/beanstalk-go/internal/constant"
)

var DrainMode int

// RandInstanceHex 生成随机实例编号
func RandInstanceHex() (string, error) {
	randData := make([]byte, constant.InstanceIDBytes)
	if _, err := rand.Read(randData); err != nil {
		return "", err
	}
	instanceHex := make([]byte, constant.InstanceIDBytes*2+1)
	for i := 0; i < constant.InstanceIDBytes; i++ {
		d := fmt.Sprintf("%02x", randData[i])
		instanceHex[i*2] = d[0]
		instanceHex[i*2+1] = d[1]
	}
	return fmt.Sprintf("%s", instanceHex), nil
}

// GetUname 获取系统信息
func GetUname() (*unix.Utsname, error) {
	utsName := &unix.Utsname{}
	if err := unix.Uname(utsName); err != nil {
		return nil, err
	}
	return utsName, nil
}
