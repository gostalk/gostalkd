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
	"flag"
	"fmt"
	"os"
	osUser "os/user"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/sjatsh/beanstalk-go/constant"
	"github.com/sjatsh/beanstalk-go/model"
)

var (
	Port                  = flag.Int("p", constant.DefaultPort, "listen on port (default is 11400)")
	BingLogDir            = flag.String("b", "", "write-ahead log directory")                                                                                                                      // binlog dir
	FsyncMs               = flag.Int64("f", constant.DefaultFsyncMs, "fsync at most once every MS milliseconds (default is 50ms);use -f0 for \"always fsync\"")                                    // fsync binlog ms
	FsyncNever            = flag.Bool("F", false, "never fsync")                                                                                                                                   // never fsync
	ListenAddr            = flag.String("l", constant.DefaultListenAddr, "listen on address (default is 0.0.0.0)")                                                                                 // server listen addr
	User                  = flag.String("u", "", "become user and group")                                                                                                                          // become user and group
	MaxJobSize            = flag.Int64("z", constant.DefaultMaxJobSize, "set the maximum job size in bytes (default is 65535);max allowed is 1073741824 bytes")                                    // max job size
	EachWriteAheadLogSize = flag.Int("s", constant.DefaultEachWriteAheadLogSize, "set the size of each write-ahead log file (default is 10485760);will be rounded up to a multiple of 4096 bytes") // each write ahead log size
	ShowVersion           = flag.Bool("v", false, "show version information")                                                                                                                      // show version
	Verbosity             = flag.Bool("V", false, "increase verbosity")                                                                                                                            // increase verbosity
	LogLevel              = flag.String("L", "warn", "set the log level, switch one in (panic, fatal, error, warn, waring, info, debug, trace)")
)

func OptParse(s *model.Server) {
	if *ShowVersion {
		fmt.Printf("beanstalk-go %s\n", Version)
		os.Exit(0)
	}

	logLevel, err := logrus.ParseLevel(*LogLevel)
	if err != nil {
		panic(err)
	}
	Log.SetLevel(logLevel)

	if *MaxJobSize > constant.JobDataSizeLimitMax {
		Log.Warnf("maximum job size was set to %d", constant.JobDataSizeLimitMax)
		*MaxJobSize = constant.JobDataSizeLimitMax
	}

	s.Wal.FileSize = *EachWriteAheadLogSize
	s.Wal.SyncRate = *FsyncMs * 1000000
	s.Wal.WantSync = 1

	if *FsyncNever {
		s.Wal.WantSync = 0
	}

	if *User != "" {
		su(*User)
	}

	if *BingLogDir != "" {
		s.Wal.Dir = *BingLogDir
		s.Wal.Use = 1
	}
}

func su(user string) {
	usr, err := osUser.Lookup(user)
	if err != nil {
		Log.Panicln(err)
	}
	gid, _ := strconv.ParseInt(usr.Gid, 10, 64)
	uid, _ := strconv.ParseInt(usr.Uid, 10, 64)
	if err := syscall.Setgid(int(gid)); err != nil {
		Log.Panicln(err)
	}
	if err := syscall.Setuid(int(uid)); err != nil {
		Log.Panicln(err)
	}
}
