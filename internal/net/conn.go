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
	"math"
	"net"
	"os"
	"time"

	"github.com/sjatsh/beanstalk-go/internal/core"
	"github.com/sjatsh/beanstalk-go/internal/structure"
	"github.com/sjatsh/beanstalk-go/internal/utils"
)

const (
	SafetyMargin = 1000000000

	ConnTypeProducer = 1 // 是生产者
	ConnTypeWorker   = 2 // 有reserve job添加WORKER类型
	ConnTypeWaiting  = 4 // 如果连接等待数据 设置WAITING类型
)

var (
	CurConnCt     = 0
	CurWorkerCt   = 0
	CurProducerCt = 0
	TotConnCt     = 0
)

func onWatch(a *structure.Ms, t interface{}, i int) {
	t.(*structure.Tube).WatchingCt++
}

func onIgnore(a *structure.Ms, t interface{}, i int) {
	t.(*structure.Tube).WatchingCt--
}

func MakeConn(f *os.File, startState int, use, watch *structure.Tube) *structure.Coon {
	c := &structure.Coon{
		ReservedJobs: core.NewJob(),
	}
	w := structure.NewMs()
	c.Watch = w

	w.WithInsertEventFn(onWatch)
	w.WithDelEventFn(onIgnore)
	w.Append(watch)

	c.Use = use
	use.UsingCt++

	c.Sock.F = f
	ln, err := net.FileListener(f)
	if err != nil {
		return nil
	}
	c.Sock.Ln = ln
	c.State = startState
	c.PendingTimeout = -1
	c.TickPos = 0
	c.InCoons = 0

	core.JobListRest(c.ReservedJobs)

	CurConnCt++
	TotConnCt++
	return c
}

func ConnDeadLineSoon(c *structure.Coon) bool {
	t := time.Now().UnixNano()
	j := ConnSoonestJob(c)
	if j == nil {
		return false
	}
	if t < j.R.DeadlineAt-SafetyMargin {
		return false
	}
	return true
}

func RemoveThisReservedJob(c *structure.Coon, j *structure.Job) *structure.Job {
	core.JobListRemove(j)
	if j != nil {
		utils.GlobalState.ReservedCt--
		j.Tube.Stat.ReservedCt--
		j.Reserver = nil
	}
	c.SoonestJob = nil
	return j
}

func ConnTimeout(c *structure.Coon) {
	shoudTimeOut := false
	if ConnWaiting(c) && ConnDeadLineSoon(c) {
		shoudTimeOut = true
	}

	for j := ConnSoonestJob(c); j != nil; j = ConnSoonestJob(c) {
		if j.R.DeadlineAt >= time.Now().UnixNano() {
			break
		}
		if j == c.OutJob {
			c.OutJob = core.JobCopy(c.OutJob)
		}
		utils.TimeoutCt++
		j.R.TimeoutCt++
		if EnqueueJob(c.Srv, RemoveThisReservedJob(c, j), 0, false) < 1 {
			core.BuryJob(c.Srv, j, false) /* out of memory, so bury it */
		}
		ConnSched(c)
	}

	if shoudTimeOut {
		RemoveWaitingCoon(c)
		ReplyMsg(c, MsgDeadlineSoon)
	} else if ConnWaiting(c) && c.PendingTimeout >= 0 {
		c.PendingTimeout = -1
		RemoveWaitingCoon(c)
		ReplyMsg(c, MsgTimedOut)
	}
}

func ConnWaiting(c *structure.Coon) bool {
	if c.Type&ConnTypeWaiting > 0 {
		return true
	}
	return false
}

func ConnSoonestJob(c *structure.Coon) *structure.Job {
	if c.SoonestJob != nil {
		return c.SoonestJob
	}

	for j := c.ReservedJobs.Next; j != c.ReservedJobs; j = j.Next {
		SetSoonestJob(c, j)
	}
	return c.SoonestJob
}

func CoonTickAt(c *structure.Coon) int64 {
	margin, shouldTimeout := int64(0), int64(0)
	t := int64(math.MaxInt64)
	if ConnWaiting(c) {
		margin = SafetyMargin
	}
	if HasReservedJob(c) {
		t = ConnSoonestJob(c).R.DeadlineAt - time.Now().UnixNano() - margin
		shouldTimeout = 1
	}
	if c.PendingTimeout >= 0 {
		t = int64(math.Min(float64(t), float64(c.PendingTimeout)*1000000000))
		shouldTimeout = 1
	}
	if shouldTimeout == 1 {
		return time.Now().UnixNano() + t
	}
	return 0
}

func HasReservedJob(c *structure.Coon) bool {
	return !core.JobListEmpty(c.ReservedJobs)
}

func EnqueueReservedJobs(c *structure.Coon) {
	for !core.JobListEmpty(c.ReservedJobs) {
		j := core.JobListRemove(c.ReservedJobs.Next)
		r := EnqueueJob(c.Srv, j, 0, false)
		if r == 0 {
			core.BuryJob(c.Srv, j, false)
		}
		utils.GlobalState.ReservedCt--
		j.Tube.Stat.ReservedCt--
		c.SoonestJob = nil
	}
}

func ConnSched(c *structure.Coon) {
	if c.InCoons > 0 {
		c.Srv.Conns.Remove(c.TickPos)
		c.InCoons = 0
	}
	c.TickAt = CoonTickAt(c)
	if c.TickAt > 0 {
		c.Srv.Conns.Push(&structure.Item{Value: c})
		c.InCoons = 1
	}
}

func ConnLess(ca, cb *structure.Item) bool {
	a := ca.Value.(*structure.Coon)
	b := cb.Value.(*structure.Coon)
	return a.TickAt < b.TickAt
}

func ConnSetPos(item *structure.Item, i int) {
	item.Value.(*structure.Coon).TickPos = i
}

func RemoveWaitingCoon(c *structure.Coon) {
	if c.Type&ConnTypeWaiting <= 0 {
		return
	}
	c.Type &= ConnTypeWaiting
	utils.GlobalState.WaitingCt--
	c.Watch.Iterator(func(item interface{}) (bool, error) {
		t := item.(*structure.Tube)
		t.Stat.WaitingCt--
		t.WaitingConns.Remove(c)
		return false, nil
	})
}

func ReserveJob(c *structure.Coon, j *structure.Job) {
	j.Tube.Stat.ReservedCt++
	j.R.ReserveCt++

	j.R.DeadlineAt = time.Now().UnixNano() + j.R.TTR
	j.R.State = core.Reserved
	core.JobListInsert(c.ReservedJobs, j)
	j.Reserver = c
	c.PendingTimeout = -1
	SetSoonestJob(c, j)
}

func SetSoonestJob(c *structure.Coon, j *structure.Job) {
	if c.SoonestJob == nil || j.R.DeadlineAt < c.SoonestJob.R.DeadlineAt {
		c.SoonestJob = j
	}
}

func ConnClose(c *structure.Coon) {

	// 连接从epoll中删除
	if _, err := SockWant(&c.Sock, 0); err != nil {
		return
	}

	if c.Sock.Ln != nil {
		c.Sock.Ln.Close()
	}
	if c.Sock.F != nil {
		c.Sock.F.Close()
	}
	c.InJob = nil

	if c.OutJob != nil && c.OutJob.R.State == core.Copy {
		c.OutJob = nil
	}

	c.InJob = nil
	c.OutJob = nil
	c.InJobRead = 0

	if c.Type&ConnTypeProducer > 0 {
		CurProducerCt--
	}
	if c.Type&ConnTypeWorker > 0 {
		CurWorkerCt--
	}

	CurConnCt--

	RemoveWaitingCoon(c)
	if HasReservedJob(c) {
		EnqueueReservedJobs(c)
	}

	c.Watch.Clear()
	c.Use.UsingCt--
	c.Use = nil

	if c.InCoons > 0 {
		c.Srv.Conns.Remove(c.TickPos)
		c.InCoons = 0
	}
	c = nil
}
