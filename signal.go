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
package main

import (
	"os"
	"sync/atomic"

	"github.com/gostalk/gostalkd/core"
	"github.com/gostalk/gostalkd/model"
	"github.com/gostalk/gostalkd/utils"
)

// sigIntTermHandle
func sigIntTermHandle(ch chan os.Signal, wal *model.Wal) {
	go func() {
		<-ch
		utils.Log.Warnln("get sigint run clean")

		utils.Log.Warnln("sync wal file start")
		core.WalSync(wal)
		utils.Log.Warnln("sync wal file stop")

		os.Exit(0)
	}()
}

// sigpipeHandle
func sigpipeHandle(ch chan os.Signal) {
	go func() {
		for _ = range ch {
		}
	}()
}

// enterDrainMode
func enterDrainMode(ch chan os.Signal) {
	go func() {
		for {
			<-ch
			utils.Log.Warnln("get SIGUSR1 sig")
			atomic.StoreInt64(&utils.DrainMode, 1)
		}
	}()
}
