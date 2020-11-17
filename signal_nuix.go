// +build linux darwin openbsd

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
	"os/signal"
	"syscall"

	"github.com/gostalk/gostalkd/model"
)

// setSigHandlers
func setSigHandlers(srv *model.Server) {
	sigpipe := make(chan os.Signal, 1)
	signal.Notify(sigpipe, syscall.SIGPIPE)
	sigpipeHandle(sigpipe)

	sigUsr1 := make(chan os.Signal, 1)
	signal.Notify(sigUsr1, syscall.SIGUSR1)
	enterDrainMode(sigUsr1)

	sigIntTerm := make(chan os.Signal, 3)
	signal.Notify(sigIntTerm, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	sigIntTermHandle(sigIntTerm, &srv.Wal)
}
