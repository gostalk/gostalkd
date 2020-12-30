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
	"hash"
	"hash/fnv"
	"math"
	"sync"
	"time"

	"github.com/gostalk/gostalkd/internal/structure/ms"
	"github.com/gostalk/gostalkd/internal/structure/xheap"
)

const defaultTubesSlotNum = 1024

type Tube struct {
	name      string
	existChan chan struct{}

	waitingClient *ms.Ms

	delayJobs *xheap.XHeap
	readyJobs *xheap.XHeap

	buried *Job
}

type tubes struct {
	pool  sync.Pool
	tubes []*tubeEntity
}

type tubeEntity struct {
	l sync.RWMutex
	m map[uint64]*Tube
}

var (
	Tubes *tubes

	defaultName = "default"
	DefaultTube = NewTube(defaultName)
)

// init init default to tubes
func init() {
	Tubes = &tubes{
		tubes: make([]*tubeEntity, defaultTubesSlotNum),
	}
	Tubes.pool.New = func() interface{} {
		return fnv.New64a()
	}
	for idx := range Tubes.tubes {
		Tubes.tubes[idx] = &tubeEntity{
			m: make(map[uint64]*Tube),
		}
	}

	h := Tubes.pool.Get().(hash.Hash64)
	h.Write([]byte(defaultName))
	sum64 := h.Sum64()
	h.Reset()
	Tubes.pool.Put(h)

	Tubes.tubes[sum64%defaultTubesSlotNum].m[sum64] = DefaultTube
}

// NewTube
func NewTube(name string) *Tube {
	t := &Tube{
		name:          name,
		waitingClient: ms.NewMs().Lock(),
		delayJobs:     xheap.NewXHeap().Lock(),
		readyJobs:     xheap.NewXHeap().Lock(),
		existChan:     make(chan struct{}),
	}
	go t.scanReadyJobs()
	return t
}

// Name
func (t *Tube) Name() string {
	return t.name
}

// GetTube
func (t *tubes) GetTube(name string) (*Tube, bool) {
	h := t.pool.Get().(hash.Hash64)
	h.Write([]byte(name))
	sum64 := h.Sum64()
	h.Reset()
	Tubes.pool.Put(h)

	tube := t.tubes[sum64%defaultTubesSlotNum]
	tube.l.RLock()
	defer tube.l.RUnlock()

	tr, ok := tube.m[sum64]
	if ok {
		return tr, true
	}
	return nil, false
}

// GetOrMakeTube
func (t *tubes) GetOrMakeTube(name string) *Tube {
	h := t.pool.Get().(hash.Hash64)
	h.Write([]byte(name))
	sum64 := h.Sum64()
	h.Reset()
	Tubes.pool.Put(h)

	tube := t.tubes[sum64%defaultTubesSlotNum]
	tr, ok := tube.m[sum64]
	if !ok {
		tube.l.Lock()
		defer tube.l.Unlock()
		tr, ok = tube.m[sum64]
		if !ok {
			tr = NewTube(name)
			tube.m[sum64] = tr
		}
	}
	return tr
}

// PushDelayJob
func (t *Tube) PushDelayJob(j *DelayJob) {
	t.delayJobs.Push(j)
}

// RemoveDelayedJob
func (t *Tube) RemoveDelayedJob(j *Job) *Job {
	if j == nil || j.R.State != Delayed {
		return nil
	}
	item := t.delayJobs.RemoveByIdx(j.heapIndex)
	if item == nil {
		return nil
	}
	return item.(*DelayJob).Job
}

// PushReadyJob
func (t *Tube) PushReadyJob(j *ReadyJob) {
	t.readyJobs.Push(j)
	t.reserveJob()
}

// RemoveReadyJob
func (t *Tube) RemoveReadyJob(j *Job) *Job {
	if j == nil || j.R.State != Ready {
		return nil
	}
	item := t.readyJobs.RemoveByIdx(j.heapIndex)
	if item == nil {
		return nil
	}
	return item.(*ReadyJob).Job
}

// RemoveBuriedJob
func (t *Tube) RemoveBuriedJob(j *Job) *Job {
	if j == nil || j.R.State != Buried {
		return nil
	}
	j = t.buried.ListRemove(j)
	if j != nil {
		// TODO 统计
	}
	return j
}

// PutReply
func (t *Tube) PutClient(c *Client) {
	t.waitingClient.Append(c)
}

// Close
func (t *Tube) Close() {
	t.existChan <- struct{}{}
}

// scanDelayJobs
func (t *Tube) scanReadyJobs() {
	period := time.Second
	timer := time.NewTimer(period)
	readyJobs := make([]xheap.HeapItem, 0)

	for {
		select {
		case <-t.existChan:
			return
		case <-timer.C:

			jobs, delayPeriod := t.findAndPopAllDelayJobs()
			readyJobs = append(readyJobs, jobs...)
			period = time.Duration(int64(math.Min(float64(period), float64(delayPeriod))))

			jobs, delayPeriod = t.soonestReservedJobs()
			readyJobs = append(readyJobs, jobs...)
			period = time.Duration(int64(math.Min(float64(period), float64(delayPeriod))))

			if len(readyJobs) > 0 {
				t.readyJobs.Push(readyJobs...)
				readyJobs = make([]xheap.HeapItem, 0, cap(readyJobs))
			}
		}

		if t.readyJobs.Len() > 0 {
			t.reserveJob()
		}
		timer.Reset(period)
	}
}

// findAndPopAllDelayJobs
func (t *Tube) findAndPopAllDelayJobs() ([]xheap.HeapItem, time.Duration) {
	period := time.Second
	readyJobs := make([]xheap.HeapItem, 0)
	for {
		delayItem := t.delayJobs.Take()
		if delayItem == nil {
			break
		}
		delayTime := delayItem.(*DelayJob).R.DeadlineAt - time.Now().UnixNano()
		if delayTime > 0 {
			period = time.Duration(int64(math.Min(float64(period), float64(delayTime))))
			break
		}

		popItem := t.delayJobs.Pop()
		if popItem == nil {
			break
		}
		if popItem != delayItem {
			continue
		}
		readyJobs = append(readyJobs, popItem.(*DelayJob).Cover2ReadyJob())
	}
	return readyJobs, period
}

// soonestReservedJobs
func (t *Tube) soonestReservedJobs() ([]xheap.HeapItem, time.Duration) {
	period := time.Second
	readyJobs := make([]xheap.HeapItem, 0)
	t.waitingClient.Iterator(func(item interface{}) bool {
		cli := item.(*Client)
		for {
			reservedJob := cli.reservedJobs.Take(0)
			if reservedJob == nil {
				break
			}
			delayTime := reservedJob.(*ReservedJob).R.DeadlineAt - time.Now().UnixNano()
			if delayTime > 0 {
				period = time.Duration(int64(math.Min(float64(period), float64(delayTime))))
				break
			}

			popItem := cli.reservedJobs.Pop()
			if popItem == nil {
				break
			}
			if popItem != reservedJob {
				continue
			}
			readyJobs = append(readyJobs, popItem.(*ReservedJob).Cover2ReadyJob())
		}
		return true
	})
	return readyJobs, period
}

// reserveJob
func (t *Tube) reserveJob() {
	if t.readyJobs.Len() <= 0 || t.waitingClient.Len() <= 0 {
		return
	}
	for t.readyJobs.Len() > 0 && t.waitingClient.Len() > 0 {
		j := t.readyJobs.Pop()
		c := t.waitingClient.Take()
		if j == nil || c == nil {
			break
		}
		readJob := j.(*ReadyJob)
		client := c.(*Client)

		client.ReplyJob(readJob.Job, MsgReserved)

		readJob.R.State = Reserved
		readJob.R.DeadlineAt = time.Now().UnixNano() + readJob.R.TTR

		// push job to client reserved heap
		client.PushReservedJob(readJob.Cover2ReservedJob())
	}
}
