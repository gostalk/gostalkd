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
	"io"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gostalk/gostalkd/internal/structure/xheap"
)

type Tube struct {
	name      string
	existChan chan struct{}

	waitingClientLock sync.RWMutex
	waitingClient     map[uint64]*Client

	delayJobs *xheap.XHeap
	readyJobs *xheap.XHeap
}

type _tubes struct {
	sync.Map
}

var (
	Tubes = _tubes{}

	defaultName = "default"
	defaultTube = NewTube(defaultName)
)

// init init default to tubes
func init() {
	Tubes.Store(defaultName, defaultTube)
}

// NewTube
func NewTube(name string) *Tube {
	t := &Tube{
		name:      name,
		delayJobs: xheap.NewXHeap().Lock(),
		readyJobs: xheap.NewXHeap().Lock(),
		existChan: make(chan struct{}),
	}
	go t.scanDelayJobs()
	return t
}

// Name
func (t *Tube) Name() string {
	return t.name
}

// GetTube
func (t *_tubes) GetTube(name string) (*Tube, bool) {
	tr, ok := t.Load(name)
	if ok {
		return tr.(*Tube), true
	}
	return nil, false
}

// GetOrMakeTube
func (t *_tubes) GetOrMakeTube(name string) *Tube {
	tr, ok := t.Load(name)
	if !ok {
		tr = NewTube(name)
		tr, _ = t.LoadOrStore(name, tr)
	}
	return tr.(*Tube)
}

// PushDelayJob
func (t *Tube) PushDelayJob(j *DelayJob) {
	t.delayJobs.Push(j)
}

// PushReadyJob
func (t *Tube) PushReadyJob(j *ReadyJob) {
	t.readyJobs.Push(j)
	t.reserveJob()
}

// PutWaitingClient
func (t *Tube) PutWaitingClient(c *Client) {
	t.waitingClientLock.Lock()
	t.waitingClient[c.session.ID()] = c
	t.waitingClientLock.Unlock()
}

// DelWaitingClient
func (t *Tube) DelWaitingClient(id uint64) bool {
	t.waitingClientLock.Lock()
	_, ok := t.waitingClient[id]
	if ok {
		delete(t.waitingClient, id)
	}
	t.waitingClientLock.Unlock()
	return ok
}

// Close
func (t *Tube) Close() {
	t.existChan <- struct{}{}
}

// scanDelayJobs
func (t *Tube) scanDelayJobs() {
	period := time.Second
	timer := time.NewTimer(period)

	for {
		select {
		case <-t.existChan:
			return
		case <-timer.C:
			readyJobs := make([]xheap.HeapItem, 0)
			for {
				delayItem := t.delayJobs.Take()
				if delayItem == nil {
					break
				}
				delayTime := delayItem.(*DelayJob).R.DeadlineAt - time.Now().UnixNano()
				if delayTime <= 0 {
					readyJobs = append(readyJobs, t.delayJobs.Pop())
				} else {
					period = time.Duration(int64(math.Min(float64(period), float64(delayTime))))
					break
				}
			}
			if len(readyJobs) > 0 {
				t.readyJobs.Push(readyJobs...)
			}
		}

		if t.readyJobs.Len() > 0 {
			t.reserveJob()
		}
		timer.Reset(period)
	}
}

// replyJob
func (t *Tube) reserveJob() {
	if len(t.waitingClient) <= 0 {
		return
	}

	replyMsg := bytes.Buffer{}
	for item := t.readyJobs.Pop(); item != nil; item = t.readyJobs.Pop() {

		j := item.(*ReadyJob)
		replyMsg.WriteString(MsgReserved)
		replyMsg.WriteString(" ")
		replyMsg.WriteString(strconv.FormatUint(j.R.ID, 10))
		replyMsg.WriteString(" ")
		replyMsg.WriteString(strconv.Itoa(len(j.body)))
		replyMsg.WriteString("\r\n")

		t.waitingClientLock.RLock()
		if len(t.waitingClient) <= 0 {
			t.waitingClientLock.RUnlock()
			return
		}

		// reply all waiting conn
		for _, c := range t.waitingClient {
			if _, err := c.GetReply().Write(replyMsg.Bytes()); err == io.EOF {
				continue
			}
			if _, err := c.GetReply().Write(j.body); err == io.EOF {
				continue
			}
			if err := c.GetReply().Flush(); err == io.EOF {
				continue
			}
		}

		t.waitingClientLock.RUnlock()

		replyMsg.Reset()
		j.R.State = JobReserved
	}
}
