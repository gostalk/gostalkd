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
package structure

import (
	"net"
	"os"
)

type Socket struct {
	F     *os.File     // 打开文件
	Ln    net.Listener // 监听器
	H     Handle       // 处理函数
	X     interface{}  // 服务器对象|Conn
	Added int16        // 是否已经添加到epoll
}

type Handle func(interface{}, byte)

type Coon struct {
	Srv            *Server
	Sock           Socket
	State          int
	Type           int
	Next           *Coon
	Use            *Tube
	TickAt         int64
	TickPos        int
	InCoons        int
	SoonestJob     *Job
	Rw             byte
	PendingTimeout int
	HalfClosed     bool

	Cmd     string
	CmdLen  int
	CmdRead int

	Reply     string
	ReplyLen  int
	ReplySent int
	ReplyBuf  string

	InJobRead int64
	InJob     *Job

	OutJob     *Job
	OutJobSent int

	Watch        *Ms
	ReservedJobs Job
}
