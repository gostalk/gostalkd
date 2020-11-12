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
	"math"
	"net"
	"os"
	"strings"

	"github.com/sjatsh/beanstalk-go/constant"
	"github.com/sjatsh/beanstalk-go/core"
	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/structure"
	"github.com/sjatsh/beanstalk-go/utils"
)

// NewServer 创建一个服务器
func NewServer(options ...model.ServerOption) (*model.Server, error) {
	opts := &model.ServerOptions{}
	for _, opt := range options {
		opt(opts)
	}
	s := &model.Server{
		Options: *opts,
		Connes:  structure.NewHeap(),
	}

	sock := &model.Socket{}
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
	if sockWant(sock, 'r') != nil {
		panic(err)
	}

	s.Sock = sock
	s.Connes.SetLessFn(connLess)
	s.Connes.SetPosFn(connSetPos)

	return s, nil
}

// Start 启动服务端并从epoll|kqueue监听socket读事件
func Start(s *model.Server) error {
	var sock *model.Socket
	for {
		period := protTick(s)
		rw, err := sockNext(&sock, period)
		if err != nil {
			os.Exit(1)
		}
		if rw != 0 {
			sock.H(sock.X, rw)
		}
	}
	return nil
}

// makeServerListener 根据配置创建unix或tcp listener
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

// srvAccept server socket读取回调
func srvAccept(i interface{}, ev byte) {
	s := i.(*model.Server)
	hAccept(s, ev)
}

// hAccept 接受客户端连接并处理
func hAccept(s *model.Server, which byte) {
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

	ct := NewConn(cf, constant.StateWantCommand, core.GetDefaultTube(), core.GetDefaultTube())
	if ct == nil {
		cf.Close()
		EpollQApply()
		return
	}

	ct.Srv = s
	ct.Sock.F = cf
	ct.Sock.X = ct
	ct.Sock.H = protHandle

	if sockWant(&ct.Sock, 'r') != nil {
		cf.Close()
	}
	EpollQApply()
}

// protHandle
func protHandle(i interface{}, b byte) {
	c := i.(*model.Coon)
	hConn(c, b)
}

// hConn
func hConn(c *model.Coon, which byte) {
	if which == 'h' {
		// 客户端连接关闭
		c.HalfClosed = true
	}

	// 维护客户端连接状态
	connProcessIO(c)

	// 读取命令并且处理
	for ; commandDataReady(c); {
		c.CmdLen = scanLineEnd(c.Cmd, c.CmdRead)
		if c.CmdLen <= 0 {
			break
		}
		dispatchCmd(c)
		fillExtraData(c)
	}

	// 判断连接关闭关闭连接
	if c.State == constant.StateClose {
		EpollQRmConn(c)
		connClose(c)
	}
	EpollQApply()
}

// connProcessIO 客户端连接状态维护
func connProcessIO(c *model.Coon) {
	switch c.State {
	case constant.StateWantCommand:
		data := make([]byte, constant.LineBufSize-c.CmdRead)
		r, err := c.Sock.F.Read(data)
		if err != nil && err != io.EOF {
			return
		}
		if r == 0 {
			c.State = constant.StateClose
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
			c.State = constant.StateWantEndLine
		}
	case constant.StateWantEndLine:
		data := make([]byte, constant.LineBufSize-c.CmdRead)
		r, err := c.Sock.F.Read(data)
		if err != nil && err != io.EOF {
			return
		}
		if err == io.EOF {
			c.State = constant.StateClose
			return
		}
		c.Cmd = append(c.Cmd, data[:r]...)
		c.CmdRead += r
		c.CmdLen = scanLineEnd(c.Cmd, c.CmdRead)
		if c.CmdLen > 0 {
			replyMsg(c, constant.MsgBadFormat)
			fillExtraData(c)
			return
		}
		if c.CmdRead == constant.LineBufSize {
			c.CmdRead = 0
		}
	case constant.StateBitbucket:
		toread := int64(math.Min(float64(c.InJobRead), float64(constant.BucketBufSize)))
		bucket := make([]byte, toread)
		r, err := c.Sock.F.Read(bucket)
		if err != nil && err != io.EOF {
			return
		}
		if err == io.EOF {
			c.State = constant.StateClose
			return
		}
		c.InJobRead -= int64(r)
		if c.InJobRead == 0 {
			reply(c, c.Reply, c.ReplyLen, constant.StateSendWord)
		}
	case constant.StateWantData:
		j := c.InJob
		data := make([]byte, j.R.BodySize-c.InJobRead)
		r, err := c.Sock.F.Read(data)
		if err != nil && err != io.EOF {
			return
		}
		if err == io.EOF {
			c.State = constant.StateClose
			return
		}
		c.InJobRead += int64(r)
		maybeEnqueueInComingJob(c)
	case constant.StateSendWord:
		replySend := c.Reply[c.ReplySent : c.ReplySent+c.ReplyLen-c.ReplySent]
		r, err := c.Sock.F.Write(replySend)
		if err != nil && err != io.EOF {
			return
		}
		if r == 0 {
			c.State = constant.StateClose
			return
		}
		c.ReplySent += int64(r)
		if c.ReplySent == c.ReplyLen {
			connWantCommand(c)
			return
		}
	case constant.StateSendJob:
		j := c.OutJob
		sendData := bytes.Buffer{}
		sendData.Write(c.Reply[c.ReplySent : c.ReplyLen-c.ReplySent])
		sendData.Write(j.Body[c.OutJobSent : j.R.BodySize-c.OutJobSent])
		r, err := c.Sock.F.Write(sendData.Bytes())
		if err != nil && err != io.EOF {
			return
		}
		if r == 0 {
			c.State = constant.StateClose
			return
		}
		c.ReplySent += int64(r)
		if c.ReplySent >= c.ReplyLen {
			c.OutJobSent += c.ReplySent - c.ReplyLen
			c.ReplySent = c.ReplyLen
		}
		if c.OutJobSent == j.R.BodySize {
			connWantCommand(c)
			return
		}
	case constant.StateWait:
		if c.HalfClosed {
			c.PendingTimeout = -1
			removeWaitingCoon(c)
			replyMsg(c, constant.MsgTimedOut)
			return
		}
	}
}

// scanLineEnd
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

// wantCommand
func wantCommand(c *model.Coon) bool {
	return c.Sock.F != nil && c.State == constant.StateWantCommand
}

// commandDataReady
func commandDataReady(c *model.Coon) bool {
	return wantCommand(c) && c.CmdRead > 0
}

// enqueueInComingJob
func enqueueInComingJob(c *model.Coon) {
	j := c.InJob

	c.InJob = nil
	c.InJobRead = 0

	// body必须以\r\n结尾
	if bytes.Index(j.Body, []byte("\r\n")) == len(j.Body)-1 {
		core.JobFree(j)
		replyMsg(c, constant.MsgExpectedCrlf)
		return
	}

	if utils.DrainMode > 0 {
		core.JobFree(j)
		replyErr(c, constant.MsgDraining)
		return
	}

	if j.WalResv > 0 {
		replyErr(c, constant.MsgInternalError)
		return
	}

	// TODO  j->walresv = walresvput(&c->srv->wal, j);
	// if j.WalResv == 0 {
	// 	replyErr(c, MsgOutOfMemory)
	// 	return
	// }

	r := enqueueJob(c.Srv, j, j.R.Delay, true)
	if r < 0 {
		replyErr(c, constant.MsgInternalError)
		return
	}

	utils.GlobalState.TotalJobsCt++
	j.Tube.Stat.TotalJobsCt++

	if r == 1 {
		replyLine(c, constant.StateSendWord, string(constant.MsgInsertedFmt), j.R.ID)
		return
	}

	// 如果报错job加入休眠队列
	core.BuryJob(c.Srv, j, false)
	replyLine(c, constant.StateSendWord, string(constant.MsgBuriedFmt), j.R.ID)
}

// maybeEnqueueInComingJob
func maybeEnqueueInComingJob(c *model.Coon) {
	j := c.InJob
	if c.InJobRead == j.R.BodySize {
		enqueueInComingJob(c)
		return
	}
	c.State = constant.StateWantData
}

// connWantCommand
func connWantCommand(c *model.Coon) {
	EpollQAdd(c, 'r')
	if c.OutJob != nil && c.OutJob.R.State == constant.Copy {
		core.JobFree(c.OutJob)
	}
	c.OutJob = nil
	c.ReplySent = 0
	c.State = constant.StateWantCommand
}
