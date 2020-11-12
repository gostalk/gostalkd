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
package model

import (
	"net"
	"os"

	"github.com/sjatsh/beanstalk-go/structure"
)

// Socket socket连接信息 以及listener接口
type Socket struct {
	F           *os.File     // 打开文件
	Ln          net.Listener // 监听器
	H           Handle       // 处理函数
	X           interface{}  // 服务器对象|Conn
	Added       bool         // 是否已经添加到epoll
	AddedFilter int16        // 监听类型
}

// 回调处理函数
type Handle func(interface{}, byte)

// Coon 连接对象
type Coon struct {
	Srv            *Server // 服务对象
	Sock           Socket  // socket连接对象
	State          int     // 连接状态
	Type           int     // 连接类型
	Next           *Coon   // 链表下一条
	Use            *Tube   // 当前连接use的tube
	TickAt         int64
	TickPos        int
	InCoons        int
	SoonestJob     *Job // 快到期job
	Rw             byte // 连接类型（'r':读 'w':写）
	PendingTimeout int  // 连接阻塞超时时间
	HalfClosed     bool // 连接断开

	Cmd     []byte // cmd命令数据
	CmdLen  int    // cmd长度
	CmdRead int    // cmd当前读取长度

	Reply     []byte // 回复给客户端的数据
	ReplyLen  int64  // 数据长度
	ReplySent int64  // 已经发送的长度
	ReplyBuf  []byte

	InJobRead int64
	InJob     *Job

	OutJob     *Job
	OutJobSent int64

	Watch        *structure.Ms
	ReservedJobs *Job
}
