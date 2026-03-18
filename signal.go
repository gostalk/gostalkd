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

package main

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/gostalk/gostalkd/core"
	"github.com/gostalk/gostalkd/model"
	"github.com/gostalk/gostalkd/utils"
)

// shutdownSignaled is set to 1 when graceful shutdown begins
var shutdownSignaled int32

// sigIntTermHandle handles SIGINT/SIGTERM for graceful shutdown.
// Steps:
//  1. Mark shutdown as signaled (stop accepting new connections)
//  2. Close the listener to stop accepting new connections
//  3. Wait for existing connections to drain (or until shutdown timeout)
//  4. Sync WAL
//  5. Exit
func sigIntTermHandle(ch chan os.Signal, srv *model.Server) {
	go func() {
		<-ch
		utils.Log.Warnln("received shutdown signal, initiating graceful shutdown")

		// Step 1: Mark server as shutting down
		atomic.StoreInt32(&shutdownSignaled, 1)

		// Step 2: Close the listener to stop accepting new connections
		if srv.Sock != nil && srv.Sock.Ln != nil {
			srv.Sock.Ln.Close()
		}

		// Step 3: Wait for existing connections to drain (with timeout)
		timeout := time.Duration(srv.ShutdownTimeout) * time.Second
		deadline := time.Now().Add(timeout)
		utils.Log.Warnf("waiting up to %v for existing connections to drain", timeout)

		// Poll until all connections are gone or timeout expires
		for time.Now().Before(deadline) {
			connCount := atomic.LoadUint64(&utils.CurConnCt)
			if connCount == 0 {
				break
			}
			utils.Log.Warnf("still waiting: %d active connections", connCount)
			time.Sleep(500 * time.Millisecond)
		}

		// Step 4: Sync WAL and exit
		remaining := atomic.LoadUint64(&utils.CurConnCt)
		if remaining > 0 {
			utils.Log.Warnf("shutdown timeout reached with %d connections still active", remaining)
		}
		utils.Log.Warnln("syncing WAL before shutdown")
		core.WalSync(&srv.Wal)
		utils.Log.Warnln("shutdown complete")
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
			utils.Log.Warnln("get SIGUSR1 sig, entering drain mode")
			atomic.StoreInt64(&utils.DrainMode, 1)
		}
	}()
}
