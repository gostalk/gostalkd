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
	"os/exec"
)

var versionCmd = `if git describe >/dev/null 2>&1
then
    git describe --tags --match=dev* | sed s/^dev// | tr - + | tr -d '\n'
    if ! git diff --quiet HEAD
    then printf +mod
    fi
else
    printf unknown
fi
`

var Version string

func init() {
	r, err := exec.Command("/bin/sh", "-c", versionCmd).Output()
	if err != nil {
		panic(err)
	}
	Version = string(r)
}
