// Copyright 2020 The Gostalkd Project Authors.
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

package proto

import (
	"bytes"
	"strconv"
	"sync"
	"time"

	"github.com/gostalk/gostalkd/internal/network"
	"github.com/gostalk/gostalkd/internal/structure/ms"
	"github.com/gostalk/gostalkd/internal/structure/xheap"
)

type CliState int32

const (
	defaultClientSlot = 1024

	SafetyMargin = time.Second

	ConnTypeProducer = 1
	ConnTypeWorker   = 2
	ConnTypeWaiting  = 4

	StateWantCommand = iota // conn expects a command from the client
	StateWantData           // conn expects a job data
	StateSendJob            // conn sends job to the client
	StateSendWord           // conn sends a line reply
	StateWait               // client awaits for the job reservation
	StateBitbucket          // conn discards content
	StateClose              // conn should be closed
	StateWantEndLine        // skip until the end of a line
)

type Clients struct {
	clients []*clientEntity
}

type clientEntity struct {
	sync.RWMutex
	m map[uint64]*Client
}

type Client struct {
	session network.Session
	reply   network.Reply
	msgPoll sync.Pool

	state      CliState
	clientType int

	use          *Tube
	watch        *ms.Ms
	reservedJobs *xheap.XHeap

	inJob *Job

	pendingTimeout int64
}

var clients *Clients

func init() {
	cls := &Clients{
		clients: make([]*clientEntity, defaultClientSlot),
	}
	for idx := range cls.clients {
		cls.clients[idx] = &clientEntity{
			m: make(map[uint64]*Client),
		}
	}
	clients = cls
}

// PutClient
func (cls *Clients) PutClient(c *Client) {
	slot := cls.clients[c.session.ID()%defaultClientSlot]
	slot.Lock()
	slot.m[c.session.ID()] = c
	slot.Unlock()
}

// DelClientByID
func (cls *Clients) DelClientByID(id uint64) {
	slot := cls.clients[id%defaultClientSlot]
	slot.Lock()
	cli, ok := slot.m[id]
	if ok {
		delete(slot.m, id)
	}
	slot.Unlock()

	for reservedJob := cli.reservedJobs.Pop(); reservedJob != nil; reservedJob = cli.reservedJobs.Pop() {
		readyJob := reservedJob.(*ReservedJob).Cover2ReadyJob()
		readyJob.tube.PushReadyJob(readyJob)
	}

	cli = nil
}

// GetClientByID
func (cls *Clients) GetClientByID(id uint64) (*Client, bool) {
	slot := cls.clients[id%defaultClientSlot]
	slot.RLock()
	cli, ok := slot.m[id]
	slot.RUnlock()
	return cli, ok
}

// NewClient
func NewClient(session network.Session, reply network.Reply) *Client {
	cli := &Client{
		session:      session,
		reply:        reply,
		use:          DefaultTube,
		watch:        ms.NewMs(DefaultTube),
		reservedJobs: xheap.NewXHeap().Lock(),
	}
	cli.msgPoll.New = func() interface{} {
		return bytes.Buffer{}
	}
	return cli
}

// DoProcess
func (c *Client) DoProcess() error {
	if c.state == StateBitbucket {
		return network.ErrReadContinue
	}
	return ioProcess(c)
}

// Use
func (c *Client) Use(tube *Tube) {
	c.use = tube
}

// GetSession
func (c *Client) GetSession() network.Session {
	return c.session
}

// GetReply
func (c *Client) GetReply() network.Reply {
	return c.reply
}

// GetState
func (c *Client) GetState() CliState {
	return c.state
}

// SetState
func (c *Client) SetState(state CliState) {
	c.state = state
}

// GetReservedJobByIdx
func (c *Client) GetReservedJobByIdx(idx int) (*Job, bool) {
	j := c.reservedJobs.Take(idx)
	if j == nil {
		return nil, false
	}
	return j.(*ReservedJob).Job, true
}

// PushReservedJob
func (c *Client) PushReservedJob(j *ReservedJob) {
	c.reservedJobs.Push(j)
}

// WatchTubeHaveReady
func (c *Client) WatchTubeHaveReady() bool {
	ready := false
	c.watch.Iterator(func(item interface{}) bool {
		if item.(*Tube).readyJobs.Len() > 0 {
			ready = true
			return false
		}
		return true
	})
	return ready
}

// EnqueueWaitingClient
func (c *Client) EnqueueWaitingClient() {
	c.clientType |= ConnTypeWaiting
	c.watch.Iterator(func(item interface{}) bool {
		t := item.(*Tube)
		t.waitingClient.Append(c)
		return true
	})
}

// WaitForJob
func (c *Client) WaitForJob(timeout int64) {
	c.state = StateWait
	c.EnqueueWaitingClient()
	c.pendingTimeout = timeout
}

// ReplyJob
func (c *Client) ReplyJob(j *Job, replyMsg string) {
	msg := c.msgPoll.Get().(bytes.Buffer)
	defer func() {
		msg.Reset()
		c.msgPoll.Put(msg)
	}()

	msg.WriteString(replyMsg)
	msg.WriteString(" ")
	msg.WriteString(strconv.FormatUint(j.R.ID, 10))
	msg.WriteString(" ")
	msg.WriteString(strconv.Itoa(len(j.body)))
	msg.WriteString("\r\n")
	msg.Write(append(j.body, []byte("\r\n")...))
	c.reply.Write(msg.Bytes())
}

// ReplyMsg
func (c *Client) ReplyMsg(replyMsg string) {
	msg := c.msgPoll.Get().(bytes.Buffer)
	defer func() {
		msg.Reset()
		c.msgPoll.Put(msg)
	}()

	msg.WriteString(replyMsg)
	msg.WriteString("\r\n")
	c.reply.Write(msg.Bytes())
}

// ReplyErr
func (c *Client) ReplyErr(err string) {
	msg := c.msgPoll.Get().(bytes.Buffer)
	defer func() {
		msg.Reset()
		c.msgPoll.Put(msg)
	}()

	msg.WriteString("server error: " + err)
	c.ReplyMsg(msg.String())
}

// SetWorker
func (c *Client) SetWorker() {
	if c.clientType&ConnTypeWorker > 0 {
		return
	}
	c.clientType |= ConnTypeWorker
}
