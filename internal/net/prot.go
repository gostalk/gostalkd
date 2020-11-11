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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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

	CmdPut              = "put "
	CmdPeekJob          = "peek "
	CmdPeekReady        = "peek-ready"
	CmdPeekDelayed      = "peek-delayed"
	CmdPeekBuried       = "peek-buried"
	CmdReserve          = "reserve"
	CmdReserveTimeout   = "reserve-with-timeout "
	CmdReserveJob       = "reserve-job "
	CmdDelete           = "delete "
	CmdRelease          = "release "
	CmdBury             = "bury "
	CmdKick             = "kick "
	CmdKickJob          = "kick-job "
	CmdTouch            = "touch "
	CmdStats            = "stats"
	CmdStatsJob         = "stats-job "
	CmdUse              = "use "
	CmdWatch            = "watch "
	CmdIgnore           = "ignore "
	CmdListTubes        = "list-tubes"
	CmdListTubeUsed     = "list-tube-used"
	CmdListTubesWatched = "list-tubes-watched"
	CmdStatsTube        = "stats-tube "
	CmdQuit             = "quit"
	CmdPauseTube        = "pause-tube"

	OpUnknown          = 0
	OpPut              = 1
	OpPeekJob          = 2
	OpReserve          = 3
	OpDelete           = 4
	OpRelease          = 5
	OpBury             = 6
	OpKick             = 7
	OpStats            = 8
	OpStatsJob         = 9
	OpPeekBuried       = 10
	OpUse              = 11
	OpWatch            = 12
	OpIgnore           = 13
	OpListTubes        = 14
	OpListTubeUsed     = 15
	OpListTubesWatched = 16
	OpStatsTube        = 17
	OpPeekReady        = 18
	OpPeekDeleadyed    = 19
	OpReserveTimeout   = 20
	OpTouch            = 21
	OpQuit             = 22
	OpPauseTube        = 23
	OpKickJob          = 24
	OpReserveJob       = 25
	TpALOps            = 26
)

var Cmd2OpMap = map[string]int{
	CmdPut:              OpPut,
	CmdPeekJob:          OpPeekJob,
	CmdPeekReady:        OpPeekReady,
	CmdPeekDelayed:      OpPeekDeleadyed,
	CmdPeekBuried:       OpPeekBuried,
	CmdReserve:          OpReserve,
	CmdReserveTimeout:   OpReserveTimeout,
	CmdReserveJob:       OpReserveJob,
	CmdDelete:           OpDelete,
	CmdRelease:          OpRelease,
	CmdBury:             OpBury,
	CmdKick:             OpKick,
	CmdKickJob:          OpKickJob,
	CmdTouch:            OpTouch,
	CmdStats:            OpStats,
	CmdStatsJob:         OpStatsJob,
	CmdUse:              OpUse,
	CmdWatch:            OpWatch,
	CmdIgnore:           OpIgnore,
	CmdListTubes:        OpListTubes,
	CmdListTubeUsed:     OpListTubeUsed,
	CmdListTubesWatched: OpListTubesWatched,
	CmdStatsTube:        OpStatsTube,
	CmdQuit:             OpQuit,
	CmdPauseTube:        OpPauseTube,
}

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
	Reply(c, m, int64(len(m)), StateSendWord)
}

func Reply(c *structure.Coon, line string, len int64, state int) {
	if c == nil {
		return
	}
	EpollqAdd(c, 'w')

	c.Reply = []byte(line)
	c.ReplyLen = len
	c.ReplySent = 0
	c.State = state
}

func ReplyErr(c *structure.Coon, err string) {
	fmt.Printf("server error: %s", err)
	ReplyMsg(c, err)
}

func replyLine(c *structure.Coon, state int, fmtStr string, params ...interface{}) {
	c.ReplyBuf = []byte(fmt.Sprintf(fmtStr, params...))
	r := len(c.ReplyBuf)
	if r >= constant.LineBufSize {
		ReplyErr(c, MsgInternalError)
	}
	Reply(c, string(c.ReplyBuf), int64(r), state)
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

func EnqueueJob(s *structure.Server, j *structure.Job, delay int64, updateStore bool) int {
	j.Reserver = nil
	if delay > 0 {
		j.R.DeadlineAt = time.Now().UnixNano() + delay
		if !j.Tube.Delay.Push(&structure.Item{Value: j}) {
			return 0
		}
		j.R.State = core.Delayed
	} else {
		if !j.Tube.Ready.Push(&structure.Item{Value: j}) {
			return 0
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
	return 1
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

func fillExtraData(c *structure.Coon) {
	if c.Sock.F == nil {
		return
	}
	if c.CmdLen == 0 {
		return
	}
	extraBytes := c.CmdRead - c.CmdLen

	jobDataBytes := 0
	if c.InJob != nil {
		jobDataBytes = int(math.Min(float64(extraBytes), float64(c.InJob.R.BodySize)))
		c.InJob.Body = c.Cmd[c.CmdLen : c.CmdLen+jobDataBytes]
		c.InJobRead = int64(jobDataBytes)
	} else if c.InJobRead > 0 {
		jobDataBytes = int(math.Min(float64(extraBytes), float64(c.InJobRead)))
		c.InJobRead -= int64(jobDataBytes)
	}

	cmdBytes := extraBytes - jobDataBytes
	c.Cmd = c.Cmd[c.CmdLen+jobDataBytes : c.CmdLen+jobDataBytes+cmdBytes]
	c.CmdRead = cmdBytes
	c.CmdLen = 0
}

func whichCmd(cmd string) int {
	for k, v := range Cmd2OpMap {
		if strings.HasPrefix(cmd, k) {
			return v
		}
	}
	return OpUnknown
}

func readInt(buf []byte) (int64, int, error) {
	begin := 0
	for i := 0; i < len(buf); i++ {
		if buf[i] != ' ' {
			break
		}
		begin++
	}
	if buf[begin] < '0' || buf[begin] > '9' {
		return 0, 0, errors.New("fmt error")
	}

	var data []byte
	end := begin
	for ; end < len(buf); end++ {
		if buf[end] == ' ' || buf[end] < '0' || buf[end] > '9' {
			break
		}
		data = append(data, buf[end])
	}

	i, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return i, end, nil
}

func skip(c *structure.Coon, n int64, msg string) {
	c.InJob = nil
	c.InJobRead = n
	fillExtraData(c)

	if c.InJobRead == 0 {
		Reply(c, msg, int64(len(msg)), StateSendWord)
		return
	}

	c.Reply = []byte(msg)
	c.ReplyLen = int64(len(msg))
	c.ReplySent = 0
	c.State = StateBitbucket
}

func readDuration(buf []byte) (int64, int, error) {
	t, next, err := readInt(buf)
	if err != nil {
		return 0, 0, err
	}
	duration := t * 1000000000
	return duration, next, nil
}

func connsetproducer(c *structure.Coon) {
	if c.Type&ConnTypeProducer > 0 {
		return
	}
	c.Type |= ConnTypeProducer
	CurProducerCt++
}

func strLen(s []byte) int {
	return bytes.Index(s, []byte("\r"))
}

func dispatchCmd(c *structure.Coon) {
	if strLen(c.Cmd) != c.CmdLen-2 {
		ReplyMsg(c, MsgBadFormat)
		return
	}

	t := whichCmd(string(c.Cmd))
	switch t {
	case OpPut:
		idx := 4
		pri, next, err := readInt(c.Cmd[idx:])
		if err != nil {
			ReplyMsg(c, MsgBadFormat)
			return
		}
		idx += next

		delay, next, err := readDuration(c.Cmd[idx:])
		if err != nil {
			ReplyMsg(c, MsgBadFormat)
			return
		}
		idx += next

		ttr, next, err := readDuration(c.Cmd[idx:])
		if err != nil {
			ReplyMsg(c, MsgBadFormat)
			return
		}
		idx += next

		bodySize, next, err := readInt(c.Cmd[idx:])
		if err != nil {
			ReplyMsg(c, MsgBadFormat)
			return
		}
		idx += next

		utils.OpCt[t]++

		if bodySize > constant.JobDataSizeLimitDefault {
			skip(c, bodySize+2, MsgJobTooBig)
			return
		}

		if c.Cmd[idx] != '\r' {
			ReplyMsg(c, MsgBadFormat)
			return
		}

		connsetproducer(c)
		if ttr < 1000000000 {
			ttr = 1000000000
		}

		c.InJob = core.MakeJobWithID(uint32(pri), delay, ttr, bodySize+2, c.Use, 0)
		fillExtraData(c)

		maybeEnqueueInComingJob(c)
	}
}
