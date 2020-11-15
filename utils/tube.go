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
package utils

import (
	"github.com/sjatsh/beanstalkd-go/constant"
)

const nameCharts = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"abcdefghijklmnopqrstuvwxyz" +
	"0123456789-+/;.$_()"

var chartsAry = [256]int32{}

func init() {
	for _, c := range nameCharts {
		chartsAry[c] = c
	}
}

// StrSpn
func StrSpn(str1 []byte) int {
	for idx, c := range str1 {
		if chartsAry[c] == 0 {
			return idx
		}
	}
	return len(str1)
}

// isValidTube
func IsValidTube(name []byte) bool {
	l := StrLen(name)
	return 0 < l && l <= constant.MaxTubeNameLen-1 &&
		StrSpn(name) == l &&
		name[0] != '-'
}
