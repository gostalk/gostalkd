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
	osUser "os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/net"
	"github.com/sjatsh/beanstalk-go/utils"
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	srv, err := net.NewServer(
		model.WithPort(*utils.Port),
		model.WithAddr(*utils.ListenAddr),
		model.WithUser(*utils.User),
	)
	if err != nil {
		panic(err)
	}

	utils.StartedAt = time.Now().UnixNano()
	utils.InstanceHex, err = utils.RandInstanceHex()
	if err != nil {
		panic(err)
	}
	utils.UtsName, err = utils.GetUname()
	if err != nil {
		panic(err)
	}

	if *utils.User != "" {
		su(*utils.User)
	}

	setSigHandlers()

	if net.Start(srv) != nil {
		panic(err)
	}
}

func su(user string) {
	usr, err := osUser.Lookup(user)
	if err != nil {
		panic(err)
	}
	gid, _ := strconv.ParseInt(usr.Gid, 10, 64)
	uid, _ := strconv.ParseInt(usr.Uid, 10, 64)
	if err := syscall.Setgid(int(gid)); err != nil {
		panic(err)
	}
	if err := syscall.Setuid(int(uid)); err != nil {
		panic(err)
	}
}

func enterDrainMode(ch chan os.Signal) {
	go func() {
		<-ch
		utils.DrainMode = 1
	}()
}

func handleSigtermPid1(ch chan os.Signal) {
	go func() {
		<-ch
		os.Exit(143)
	}()
}

func setSigHandlers() {
	sigpipe := make(chan os.Signal, 1)
	signal.Notify(sigpipe, syscall.SIGPIPE)

	sigusr := make(chan os.Signal)
	signal.Notify(sigusr, syscall.SIGUSR1)
	enterDrainMode(sigusr)

	if os.Getpid() == 1 {
		sigterm := make(chan os.Signal)
		signal.Notify(sigusr, syscall.SIGTERM)
		handleSigtermPid1(sigterm)
	}
}
