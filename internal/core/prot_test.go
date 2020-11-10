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
	"math/rand"
	"testing"
	"time"
)

func TestRandInstanceHex(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	instanceHex, err := RandInstanceHex()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(instanceHex)
}

func TestGetUname(t *testing.T) {
	utsname, err := GetUname()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(`
sysname:%s
nodename:%s
release:%s
version:%s
machine:%s
`, utsname.Sysname, utsname.Nodename, utsname.Release, utsname.Version, utsname.Machine)
}
