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
	"fmt"
	"math/rand"

	"github.com/gostalk/gostalkd/constant"
)

// SysName
type SysInfo struct {
	SysName  string
	NodeName string
	Release  string
	Version  string
	Machine  string
}

// Rusage
type Rusage struct {
	Usec  int64
	Ssec  int64
	Total int64
}

var (
	DrainMode   int64
	InstanceHex string
)

func init() {
	randData := make([]byte, constant.InstanceIDBytes)
	if _, err := rand.Read(randData); err != nil {
		panic(err)
	}
	instanceHex := make([]byte, constant.InstanceIDBytes*2+1)
	for i := 0; i < constant.InstanceIDBytes; i++ {
		d := fmt.Sprintf("%02x", randData[i])
		instanceHex[i*2] = d[0]
		instanceHex[i*2+1] = d[1]
	}
	InstanceHex = fmt.Sprintf("%s", instanceHex)
}
