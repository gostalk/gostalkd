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

	"github.com/sjatsh/beanstalk-go/constant"
	"github.com/sjatsh/beanstalk-go/core"
	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/structure"
	"github.com/sjatsh/beanstalk-go/utils"
)

// MakeConn
func NewConn(f *os.File, startState int, use, watch *model.Tube) *model.Coon {
	c := &model.Coon{
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

	utils.CurConnCt++
	utils.TotalConnCt++
	return c
}

// onWatch
func onWatch(a *structure.Ms, t interface{}, i int) {
	t.(*model.Tube).WatchingCt++
}

// onIgnore
func onIgnore(a *structure.Ms, t interface{}, i int) {
	t.(*model.Tube).WatchingCt--
}

// connDeadLineSoon
func connDeadLineSoon(c *model.Coon) bool {
	t := time.Now().UnixNano()
	j := connSoonestJob(c)
	if j == nil {
		return false
	}
	if t < j.R.DeadlineAt-constant.SafetyMargin {
		return false
	}
	return true
}

// connReady
func connReady(c *model.Coon) bool {
	ready := false
	c.Watch.Iterator(func(item interface{}) (bool, error) {
		if item.(*model.Tube).Ready.Len() > 0 {
			ready = true
			return true, nil
		}
		return false, nil
	})
	return ready
}

func removeThisReservedJob(c *model.Coon, j *model.Job) *model.Job {
	core.JobListRemove(j)
	if j != nil {
		utils.GlobalState.ReservedCt--
		j.Tube.Stat.ReservedCt--
		j.Reservoir = nil
	}
	c.SoonestJob = nil
	return j
}

func connTimeout(c *model.Coon) {
	shouldTimeOut := false
	if connWaiting(c) && connDeadLineSoon(c) {
		shouldTimeOut = true
	}

	for j := connSoonestJob(c); j != nil; j = connSoonestJob(c) {
		if j.R.DeadlineAt >= time.Now().UnixNano() {
			break
		}
		if j == c.OutJob {
			c.OutJob = core.JobCopy(c.OutJob)
		}
		utils.TimeoutCt++
		j.R.TimeoutCt++
		if enqueueJob(c.Srv, removeThisReservedJob(c, j), 0, false) < 1 {
			core.BuryJob(c.Srv, j, false) /* out of memory, so bury it */
		}
		connSched(c)
	}

	if shouldTimeOut {
		removeWaitingCoon(c)
		replyMsg(c, constant.MsgDeadlineSoon)
	} else if connWaiting(c) && c.PendingTimeout >= 0 {
		c.PendingTimeout = -1
		removeWaitingCoon(c)
		replyMsg(c, constant.MsgTimedOut)
	}
}

func connWaiting(c *model.Coon) bool {
	if c.Type&constant.ConnTypeWaiting > 0 {
		return true
	}
	return false
}

func connSoonestJob(c *model.Coon) *model.Job {
	if c.SoonestJob != nil {
		return c.SoonestJob
	}
	for j := c.ReservedJobs.Next; j != c.ReservedJobs; j = j.Next {
		setSoonestJob(c, j)
	}
	return c.SoonestJob
}

func coonTickAt(c *model.Coon) int64 {
	margin, shouldTimeout := int64(0), int64(0)
	t := int64(math.MaxInt64)
	if connWaiting(c) {
		margin = constant.SafetyMargin
	}
	if hasReservedJob(c) {
		t = connSoonestJob(c).R.DeadlineAt - time.Now().UnixNano() - margin
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

func hasReservedJob(c *model.Coon) bool {
	return !core.JobListEmpty(c.ReservedJobs)
}

func enqueueReservedJobs(c *model.Coon) {
	for !core.JobListEmpty(c.ReservedJobs) {
		j := core.JobListRemove(c.ReservedJobs.Next)
		r := enqueueJob(c.Srv, j, 0, false)
		if r == 0 {
			core.BuryJob(c.Srv, j, false)
		}
		utils.GlobalState.ReservedCt--
		j.Tube.Stat.ReservedCt--
		c.SoonestJob = nil
	}
}

func connSched(c *model.Coon) {
	if c.InCoons > 0 {
		c.Srv.Connes.Remove(c.TickPos)
		c.InCoons = 0
	}
	c.TickAt = coonTickAt(c)
	if c.TickAt > 0 {
		c.Srv.Connes.Push(&structure.Item{Value: c})
		c.InCoons = 1
	}
}

func connLess(ca, cb *structure.Item) bool {
	a := ca.Value.(*model.Coon)
	b := cb.Value.(*model.Coon)
	return a.TickAt < b.TickAt
}

func connSetPos(item *structure.Item, i int) {
	item.Value.(*model.Coon).TickPos = i
}

func removeWaitingCoon(c *model.Coon) {
	if c.Type&constant.ConnTypeWaiting <= 0 {
		return
	}
	c.Type &= ^constant.ConnTypeWaiting
	utils.GlobalState.WaitingCt--
	c.Watch.Iterator(func(item interface{}) (bool, error) {
		t := item.(*model.Tube)
		t.Stat.WaitingCt--
		t.WaitingConns.Remove(c)
		return false, nil
	})
}

func reserveJob(c *model.Coon, j *model.Job) {
	j.Tube.Stat.ReservedCt++
	j.R.ReserveCt++

	j.R.DeadlineAt = time.Now().UnixNano() + j.R.TTR
	j.R.State = constant.Reserved
	core.JobListInsert(c.ReservedJobs, j)
	j.Reservoir = c
	c.PendingTimeout = -1
	setSoonestJob(c, j)
}

func setSoonestJob(c *model.Coon, j *model.Job) {
	if c.SoonestJob == nil || j.R.DeadlineAt < c.SoonestJob.R.DeadlineAt {
		c.SoonestJob = j
	}
}

func connClose(c *model.Coon) {

	// 连接从epoll中删除
	if sockWant(&c.Sock, 0) != nil {
		return
	}

	if c.Sock.Ln != nil {
		c.Sock.Ln.Close()
	}
	if c.Sock.F != nil {
		c.Sock.F.Close()
	}
	c.InJob = nil

	if c.OutJob != nil && c.OutJob.R.State == constant.Copy {
		c.OutJob = nil
	}

	c.InJob = nil
	c.OutJob = nil
	c.InJobRead = 0

	if c.Type&constant.ConnTypeProducer > 0 {
		utils.CurProducerCt--
	}
	if c.Type&constant.ConnTypeWorker > 0 {
		utils.CurWorkerCt--
	}

	utils.CurConnCt--

	removeWaitingCoon(c)
	if hasReservedJob(c) {
		enqueueReservedJobs(c)
	}

	c.Watch.Clear()

	c.Use.UsingCt--
	c.Use = nil

	if c.InCoons > 0 {
		c.Srv.Connes.Remove(c.TickPos)
		c.InCoons = 0
	}
	c = nil
}
