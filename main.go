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

	"github.com/sjatsh/beanstalk-go/internal/constant"
	"github.com/sjatsh/beanstalk-go/internal/core"
	"github.com/sjatsh/beanstalk-go/internal/net"
	"github.com/sjatsh/beanstalk-go/internal/structure"
	"github.com/sjatsh/beanstalk-go/internal/utils"
)

var (
	port                  = flag.Int("p", constant.DefaultPort, "listen on port (default is 11400)")
	bingLogDir            = flag.String("b", "", "write-ahead log directory")                                                                                                                      // binlog dir
	fsyncMs               = flag.Int("f", constant.DefaultFsyncMs, "fsync at most once every MS milliseconds (default is 50ms);use -f0 for \"always fsync\"")                                      // fsync binlog ms
	fsyncNever            = flag.Bool("F", false, "never fsync")                                                                                                                                   // never fsync
	listenAddr            = flag.String("l", constant.DefaultListenAddr, "listen on address (default is 0.0.0.0)")                                                                                 // server listen addr
	user                  = flag.String("u", "", "become user and group")                                                                                                                          // become user and group
	maxJobSize            = flag.Int("z", constant.DefaultMaxJobSize, "set the maximum job size in bytes (default is 65535);max allowed is 1073741824 bytes")                                      // max job size
	eachWriteAheadLogSize = flag.Int("s", constant.DefaultEachWriteAheadLogSize, "set the size of each write-ahead log file (default is 10485760);will be rounded up to a multiple of 4096 bytes") // each write ahead log size
	showVersion           = flag.Bool("v", false, "show version information")                                                                                                                      // show version
	verbosity             = flag.Bool("V", false, "increase verbosity")                                                                                                                            // increase verbosity
)

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
		core.DrainMode = 1
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

var srv *structure.Server

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	var err error
	srv, err = net.NewServer(
		net.WithPort(*port),
		net.WithAddr(*listenAddr),
		net.WithUser(*user),
	)
	if err != nil {
		panic(err)
	}

	utils.StartedAt = time.Now().UnixNano()
	utils.InstanceHex, err = core.RandInstanceHex()
	if err != nil {
		panic(err)
	}
	utils.UtsName, err = core.GetUname()
	if err != nil {
		panic(err)
	}

	if *user != "" {
		su(*user)
	}

	setSigHandlers()

	if err := net.Start(srv); err != nil {
		panic(err)
	}
}
