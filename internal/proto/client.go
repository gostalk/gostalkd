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
	"sync"

	"github.com/gostalk/gostalkd/internal/network"
)

const defaultClientSlot = 1024

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
	state   CliState

	use          *Tube
	watch        map[string]*Tube
	reservedJobs map[uint64]*Job

	inJob *Job
}

// NewClients
func NewClients() *Clients {
	cls := &Clients{
		clients: make([]*clientEntity, defaultClientSlot),
	}
	for idx := range cls.clients {
		cls.clients[idx] = &clientEntity{
			m: make(map[uint64]*Client),
		}
	}
	return cls
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
	delete(slot.m, id)
	slot.Unlock()
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
	return &Client{
		session: session,
		reply:   reply,
		use:     defaultTube,
		watch: map[string]*Tube{
			defaultName: defaultTube,
		},
	}
}

// ProcessData
func (c *Client) Process() error {
	if c.state == StateBitbucket {
		return nil
	}
	return ioProcess(c)
}

// GetSessionID
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

// AddWatch
func (c *Client) AddWatch(tube *Tube) {
	c.watch[tube.Name()] = tube
}

// GetWatchTubeByName
func (c *Client) GetWatchTubeByName(name string) (*Tube, bool) {
	t, ok := c.watch[name]
	return t, ok
}

// Use
func (c *Client) Use(tube *Tube) {
	c.use = tube
}

// GetUseTube
func (c *Client) GetUseTube() *Tube {
	t := c.use
	return t
}

// GetReservedJobByID
func (c *Client) GetReservedJobByID(id uint64) (*Job, bool) {
	j, ok := c.reservedJobs[id]
	return j, ok
}

// AddReservedJob
func (c *Client) AddReservedJob(j *Job) {
	c.reservedJobs[j.R.ID] = j
}

// DelReservedJob
func (c *Client) DelReservedJob(j *Job) bool {
	_, ok := c.reservedJobs[j.R.ID]
	if ok {
		delete(c.reservedJobs, j.R.ID)
	}
	return ok
}
