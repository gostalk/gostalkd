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
package net

import (
	"bytes"
	"fmt"
	"io"
	net "net"
	"os"
	"strings"

	"github.com/sjatsh/beanstalk-go/internal/constant"
	"github.com/sjatsh/beanstalk-go/internal/core"
	"github.com/sjatsh/beanstalk-go/internal/structure"
)

func makeServerListener(addr string, port int) (net.Listener, error) {
	if strings.HasPrefix(addr, "unix:") {
		a, err := net.ResolveUnixAddr("unix", strings.TrimPrefix(addr, "unix:"))
		if err != nil {
			return nil, err
		}
		l, err := net.ListenUnix("unix", a)
		if err != nil {
			return nil, err
		}
		return l, nil
	}

	a, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP("tcp", a)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func srvAccept(i interface{}, ev byte) {
	s := i.(*structure.Server)
	HAccept(s, ev)
}

// NewServer 创建一个服务器
func NewServer(options ...structure.ServerOption) (*structure.Server, error) {
	opts := &structure.ServerOptions{}
	for _, opt := range options {
		opt(opts)
	}
	s := &structure.Server{
		Options: *opts,
		Conns:   structure.NewHeap(),
	}

	sock := &structure.Socket{}
	l, err := makeServerListener(opts.Addr, opts.Port)
	if err != nil {
		return nil, err
	}

	var f *os.File
	switch nl := l.(type) {
	case *net.TCPListener:
		f, _ = nl.File()
	case *net.UnixListener:
		f, _ = nl.File()
	}

	sock.F = f
	sock.Ln = l
	sock.X = s
	sock.H = srvAccept
	if _, err := SockWant(sock, 'r'); err != nil {
		panic(err)
	}

	s.Sock = sock
	s.Conns.SetLessFn(ConnLess)
	s.Conns.SetPosFn(ConnSetPos)

	return s, nil
}

// 接受客户端连接
func HAccept(s *structure.Server, which byte) {
	c, err := s.Sock.Ln.Accept()
	if err != nil {
		return
	}

	var cf *os.File
	switch s := c.(type) {
	case *net.TCPConn:
		cf, err = s.File()
		if err != nil {
			return
		}
	case *net.UnixConn:
		cf, err = s.File()
		if err != nil {
			return
		}
	default:
		return
	}

	if int(cf.Fd()) == -1 {
		EpollQApply()
		return
	}

	ct := MakeConn(cf, StateWantCommand, core.GetDefaultTube(), core.GetDefaultTube())
	if ct == nil {
		cf.Close()
		EpollQApply()
		return
	}

	ct.Srv = s
	ct.Sock.F = cf
	ct.Sock.X = ct
	ct.Sock.H = protHandle

	if r, err := SockWant(&ct.Sock, 'r'); r == -1 || err != nil {
		cf.Close()
	}

	EpollQApply()
}

func protHandle(i interface{}, b byte) {
	c := i.(*structure.Coon)
	hConn(c, b)
}

func scanLineEnd(s []byte, size int) int {
	l := len(s)
	if l < 2 || l != size {
		return 0
	}
	idx := bytes.Index(s, []byte("\r"))
	if idx == -1 {
		return 0
	}
	if idx >= l-1 {
		return 0
	}
	if s[idx+1] == '\n' {
		return idx + 2
	}
	return 0
}

func wantCommand(c *structure.Coon) bool {
	return c.Sock.F != nil && c.State == StateWantCommand
}

func commandDataReady(c *structure.Coon) bool {
	return wantCommand(c) && c.CmdRead > 0
}

func hConn(c *structure.Coon, which byte) {
	if which == 'h' {
		// 客户端连接关闭
		c.HalfClosed = true
	}

	// 维护客户端连接状态
	connProcessIO(c)

	// 读取命令并且处理
	for ; commandDataReady(c) && c.CmdLen == scanLineEnd(c.Cmd, c.CmdRead); {
		// dispatch_cmd(c)
		// fill_extra_data(c)
	}

	// 判断连接关闭关闭连接
	if c.State == StateClose {
		EpollQRmConn(c)
		ConnClose(c)
	}
	EpollQApply()
}

func connProcessIO(c *structure.Coon) {
	switch c.State {
	case StateWantCommand:
		data := make([]byte, constant.LineBufSize-c.CmdRead)
		r, err := c.Sock.F.Read(data)
		if r == -1 || (err != nil && err != io.EOF) {
			return
		}
		if r == 0 {
			// 连接关闭
			c.State = StateClose
			return
		}
		c.Cmd = append(c.Cmd, data[:r]...)
		c.CmdRead += r
		c.CmdLen = scanLineEnd(c.Cmd, c.CmdRead)
		if c.CmdLen > 0 {
			// 读取到完整命令直接return
			return
		}
		if c.CmdRead == constant.LineBufSize {
			c.CmdRead = 0
			c.State = StateWantEndLine
		}
		return
	case StateWantEndLine:
		data := make([]byte, constant.LineBufSize-c.CmdRead)
		r, err := c.Sock.F.Read(data)
		if r == -1 || (err != nil && err != io.EOF) {
			return
		}
		if r == 0 {
			// 连接关闭
			c.State = StateClose
			return
		}
		c.Cmd = append(c.Cmd, data[:r]...)
		c.CmdRead += r
		c.CmdLen = scanLineEnd(c.Cmd, c.CmdRead)
		if c.CmdLen > 0 {
			ReplyMsg(c, MsgBadFormat)
			// TODO 填充额外数据
			return
		}
		if c.CmdRead == constant.LineBufSize {
			c.CmdRead = 0
		}
		return
	case StateBitbucket:
	case StateWantData:
	case StateSendWord:
	case StateSendJob:
	case StateWait:

	}
}

// Start 开启服务
func Start(s *structure.Server) error {
	var sock *structure.Socket
	for {
		period := ProtTick(s)
		rw, err := SockNext(&sock, period)
		if int(rw) == -1 || err != nil {
			// TODO error log
			os.Exit(1)
		}
		if rw > 0 {
			sock.H(sock.X, rw)
		}
	}
	return nil
}

// WithPort 设置端口
func WithPort(port int) structure.ServerOption {
	return func(o *structure.ServerOptions) {
		o.Port = port
	}
}

// WithAddr 设置服务端监听ip
func WithAddr(addr string) structure.ServerOption {
	return func(o *structure.ServerOptions) {
		o.Addr = addr
	}
}

// WithUser 运行用户
func WithUser(user string) structure.ServerOption {
	return func(o *structure.ServerOptions) {
		o.User = user
	}
}
