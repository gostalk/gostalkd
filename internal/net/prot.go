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
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/sjatsh/beanstalk-go/internal/constant"
	"github.com/sjatsh/beanstalk-go/internal/core"
	"github.com/sjatsh/beanstalk-go/internal/structure"
	"github.com/sjatsh/beanstalk-go/internal/utils"
)

const (
	StateWantCommand = 0 // conn expects a command from the client
	StateWantData    = 1 // conn expects a job data
	StateSendJob     = 2 // conn sends job to the client
	StateSendWord    = 3 // conn sends a line reply
	StateWait        = 4 // client awaits for the job reservation
	StateBitbucket   = 5 // conn discards content
	StateClose       = 6 // conn should be closed
	StateWantEndLine = 7 // skip until the end of a line

	MsgNotFound       = "NOT_FOUND\r\n"
	MSsgFound         = "FOUND"
	MsgReserved       = "RESERVED"
	MsgDeadlineSoon   = "DEADLINE_SOON\r\n"
	MsgTimedOut       = "TIMED_OUT\r\n"
	MsgDeleted        = "DELETED\r\n"
	MsgReleased       = "RELEASED\r\n"
	MsgBuried         = "BURIED\r\n"
	MsgKicked         = "KICKED\r\n"
	MsgTouched        = "TOUCHED\r\n"
	MsgBuriedFmt      = "BURIED %d\r\n"
	MsgInsertedFmt    = "INSERTED %d\r\n"
	MsgNotIgnored     = "NOT_IGNORED\r\n"
	MsgOutOfMemory    = "OUT_OF_MEMORY\r\n"
	MsgInternalError  = "INTERNAL_ERROR\r\n"
	MsgDraining       = "DRAINING\r\n"
	MsgBadFormat      = "BAD_FORMAT\r\n"
	MsgUnknownCommand = "UNKNOWN_COMMAND\r\n"
	MsgExpectedCrlf   = "EXPECTED_CRLF\r\n"
	MsgJobTooBig      = "JOB_TOO_BIG\r\n"
)

var epollQ *structure.Coon

func EpollqAdd(c *structure.Coon, rw byte) {
	c.Rw = rw
	ConnSched(c)
	c.Next = epollQ
	epollQ = c
}

func EpollQApply() {
	var c *structure.Coon
	for ; epollQ != nil; {
		c = epollQ
		epollQ = epollQ.Next
		c.Next = nil
		if n, err := SockWant(&c.Sock, c.Rw); n == -1 || err != nil {
			// TODO err log
			ConnClose(c)
		}
	}
}

func EpollQRmConn(c *structure.Coon) {
	var x, newhead *structure.Coon
	for ; epollQ != nil; {
		x = epollQ
		epollQ = epollQ.Next
		x.Next = nil

		if x != c {
			x.Next = newhead
			newhead = x
		}
	}
	epollQ = newhead
}

func ReplyMsg(c *structure.Coon, m string) {
	Reply(c, m, len(m), StateSendWord)
}

func Reply(c *structure.Coon, line string, len, state int) {
	if c == nil {
		return
	}
	EpollqAdd(c, 'w')

	c.Reply = line
	c.ReplyLen = len
	c.ReplySent = 0
	c.State = state
}

func ReplyErr(c *structure.Coon, err string) {
	fmt.Printf("server error: %s", err)
	ReplyMsg(c, err)
}

func replyLine(c *structure.Coon, state int, fmtStr string, msg string, id uint64, size uint64) {
	c.ReplyBuf = fmt.Sprintf(fmtStr, msg, id, size)
	r := len(c.ReplyBuf)
	if r >= constant.LineBufSize {
		ReplyErr(c, MsgInternalError)
	}
	Reply(c, c.ReplyBuf, r, state)
}

func ReplyJob(c *structure.Coon, j *structure.Job, msg string) {
	c.OutJob = j
	c.OutJobSent = 0
	replyLine(c, StateSendJob, "%s %d %d\r\n",
		msg, j.R.ID, j.R.BodySize-2)
}

func ProcessQueue() {
	now := time.Now().UnixNano()
	for j := core.NextAwaitedJob(now); j != nil; j = core.NextAwaitedJob(now) {
		j.Tube.Ready.Remove(j.HeapIndex)
		atomic.AddInt64(&utils.ReadyCt, -1)
		if j.R.Pri < constant.UrgentThreshold {
			utils.GlobalState.UrgentCt--
			j.Tube.Stat.UrgentCt--
		}

		ci := j.Tube.WaitingConns.Take()
		if ci == nil {
			continue
		}
		c := ci.(*structure.Coon)
		utils.GlobalState.ReservedCt++

		RemoveWaitingCoon(c)
		ReserveJob(c, j)
		ReplyJob(c, j, MsgReserved)
	}
}

func EnqueueJob(s *structure.Server, j *structure.Job, delay int64, updateStore bool) bool {
	j.Reserver = nil
	if delay > 0 {
		j.R.DeadlineAt = time.Now().UnixNano() + delay
		if !j.Tube.Delay.Push(&structure.Item{Value: j}) {
			return false
		}
		j.R.State = core.Delayed
	} else {
		if !j.Tube.Ready.Push(&structure.Item{Value: j}) {
			return false
		}
		j.R.State = core.Ready
		utils.ReadyCt++
		if j.R.Pri < constant.UrgentThreshold {
			utils.GlobalState.UrgentCt++
			j.Tube.Stat.UrgentCt++
		}
	}

	if updateStore {
		// TODO 写入binlog
	}

	ProcessQueue()
	return true
}

func ProtTick(s *structure.Server) time.Duration {
	var (
		d      int64
		period = constant.DefaultPeriod
		now    = time.Now().UnixNano()
	)

	for j := core.SoonestDelayedJob(); j != nil; j = core.SoonestDelayedJob() {
		d = j.R.DeadlineAt - now
		if d > 0 {
			period = int64(math.Min(float64(period), float64(d)))
			break
		}
		j.Tube.Delay.Remove(j.HeapIndex)

		// TODO 如何判断内存不足
	}

	core.GetTubes().Iterator(func(item interface{}) (bool, error) {
		t := item.(*structure.Tube)
		d := t.UnpauseAt - now
		if t.Pause > 0 && d <= 0 {
			t.Pause = 0
			ProcessQueue()
		} else if d > 0 {
			period = int64(math.Min(float64(period), float64(d)))
		}
		return false, nil
	})

	for i := s.Conns.Len(); i > s.Conns.Len(); i = s.Conns.Len() {
		ci := s.Conns.Take()
		if ci == nil {
			continue
		}
		c := ci.Value.(*structure.Coon)
		d = c.TickAt - now
		if d > 0 {
			period = int64(math.Min(float64(period), float64(d)))
			break
		}

		s.Conns.Remove(0)
		c.InCoons = 0
		ConnTimeout(c)
	}

	EpollQApply()

	return time.Duration(period)
}
