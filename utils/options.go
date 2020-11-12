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

	"github.com/sjatsh/beanstalk-go/constant"
)

var (
	Port                  = flag.Int("p", constant.DefaultPort, "listen on port (default is 11400)")
	BingLogDir            = flag.String("b", "", "write-ahead log directory")                                                                                                                      // binlog dir
	FsyncMs               = flag.Int("f", constant.DefaultFsyncMs, "fsync at most once every MS milliseconds (default is 50ms);use -f0 for \"always fsync\"")                                      // fsync binlog ms
	FsyncNever            = flag.Bool("F", false, "never fsync")                                                                                                                                   // never fsync
	ListenAddr            = flag.String("l", constant.DefaultListenAddr, "listen on address (default is 0.0.0.0)")                                                                                 // server listen addr
	User                  = flag.String("u", "", "become user and group")                                                                                                                          // become user and group
	MaxJobSize            = flag.Int("z", constant.DefaultMaxJobSize, "set the maximum job size in bytes (default is 65535);max allowed is 1073741824 bytes")                                      // max job size
	EachWriteAheadLogSize = flag.Int("s", constant.DefaultEachWriteAheadLogSize, "set the size of each write-ahead log file (default is 10485760);will be rounded up to a multiple of 4096 bytes") // each write ahead log size
	ShowVersion           = flag.Bool("v", false, "show version information")                                                                                                                      // show version
	Verbosity             = flag.Bool("V", false, "increase verbosity")                                                                                                                            // increase verbosity
)
