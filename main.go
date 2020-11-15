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
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sjatsh/beanstalk-go/core"
	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/net"
	"github.com/sjatsh/beanstalk-go/utils"
)

var version string

func init() {
	utils.Version = version
}

var srv *model.Server

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	var err error
	// new server object
	srv, err = net.NewServer(
		model.WithPort(*utils.Port),
		model.WithAddr(*utils.ListenAddr),
		model.WithUser(*utils.User),
	)
	if err != nil {
		utils.Log.Panicln(err)
	}

	// parse options
	utils.OptParse(srv)
	// sig handlers
	setSigHandlers()
	// srvAcquireWal
	net.SrvAcquireWal(srv)

	if net.Start(srv) != nil {
		utils.Log.Panicln(err)
	}
}

func setSigHandlers() {
	sigpipe := make(chan os.Signal, 1)
	signal.Notify(sigpipe, syscall.SIGPIPE)
	sigpipeHandle(sigpipe)

	sigUsr1 := make(chan os.Signal, 1)
	signal.Notify(sigUsr1, syscall.SIGUSR1)
	enterDrainMode(sigUsr1)

	sigIntTerm := make(chan os.Signal, 2)
	signal.Notify(sigIntTerm, syscall.SIGINT, syscall.SIGTERM)
	sigIntTermHandle(sigIntTerm)
}

func sigIntTermHandle(ch chan os.Signal) {
	go func() {
		<-ch
		utils.Log.Warnln("get sigint")
		core.WalMain(&srv.Wal)
		os.Exit(0)
	}()
}

func sigpipeHandle(ch chan os.Signal) {
	go func() {
		for _ = range ch {
		}
	}()
}

func enterDrainMode(ch chan os.Signal) {
	go func() {
		for {
			<-ch
			utils.Log.Warnln("get SIGUSR1 sig")
			utils.DrainMode = 1
		}
	}()
}

func handleSigtermPid1(ch chan os.Signal) {
	go func() {
		<-ch
		utils.Log.Infoln("get SIGTERM sig")
		core.WalMain(&srv.Wal)
		os.Exit(143)
	}()
}
