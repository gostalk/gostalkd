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
	"strings"
	"sync/atomic"
	"time"

	"github.com/sjatsh/beanstalk-go/constant"
	"github.com/sjatsh/beanstalk-go/core"
	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/structure"
	"github.com/sjatsh/beanstalk-go/utils"
)

var epollQ *model.Coon

// EpollQAdd
func EpollQAdd(c *model.Coon, rw byte) {
	c.Rw = rw
	connSched(c)
	c.Next = epollQ
	epollQ = c
}

//  EpollQApply
func EpollQApply() {
	var c *model.Coon
	for ; epollQ != nil; {
		c = epollQ
		epollQ = epollQ.Next
		c.Next = nil
		if sockWant(&c.Sock, c.Rw) != nil {
			connClose(c)
		}
	}
}

// EpollQRmConn
func EpollQRmConn(c *model.Coon) {
	var x, newhead *model.Coon
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

// protTick 查询最快到期job延时时间，没有默认延时1小时
func protTick(s *model.Server) time.Duration {
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
		if enqueueJob(s, j, 0, false) < 1 {
			// TODO 如何判断内存不足
		}
	}

	core.GetTubes().Iterator(func(item interface{}) (bool, error) {
		t := item.(*model.Tube)
		d := t.UnpauseAt - now
		if t.Pause > 0 && d <= 0 {
			t.Pause = 0
			processQueue()
		} else if d > 0 {
			period = int64(math.Min(float64(period), float64(d)))
		}
		return false, nil
	})

	// 从coons中找出reserve job ttr时间最小的
	for i := s.Connes.Len(); i > s.Connes.Len(); i = s.Connes.Len() {
		ci := s.Connes.Take()
		if ci == nil {
			continue
		}
		c := ci.Value.(*model.Coon)
		d = c.TickAt - now
		if d > 0 {
			period = int64(math.Min(float64(period), float64(d)))
			break
		}

		s.Connes.Remove(0)
		c.InCoons = 0
		connTimeout(c)
	}
	EpollQApply()
	return time.Duration(period)
}

// whichCmd 验证命令是否有效
func whichCmd(cmd string) int {
	for k, v := range constant.Cmd2OpMap {
		if strings.HasPrefix(cmd, k) {
			return v
		}
	}
	return constant.OpUnknown
}

// dispatchCmd 客户端命令处理
func dispatchCmd(c *model.Coon) {

	var timeout int64 = -1

	if utils.StrLen(c.Cmd) != c.CmdLen-2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	t := whichCmd(string(c.Cmd))
	switch t {
	case constant.OpPut:
		displayOpPut(c)
	case constant.OpPeekReady:
		displayOpPeekReady(c)
	case constant.OpPeekDelayed:
		displayOpPeekDelayed(c)
	case constant.OpPeekBuried:
		displayOpPeekBuried(c)
	case constant.OpPeekJob:
		displayOpPeekJob(c)
	case constant.OpReserveTimeout:
		timeout = displayOpReserveTimeout(c)
		fallthrough
	case constant.OpReserve:
		displayOpReserve(c, timeout)
	case constant.OpReserveJob:
		displayOpReserveJob(c)
	case constant.OpDelete:
		displayOpDelete(c)
	case constant.OpRelease:
		displayOpRelease(c)
	case constant.OpBury:
		displayOpBury(c)
	case constant.OpKick:
		displayOpKick(c)
	case constant.OpKickJob:
	case constant.OpTouch:
	case constant.OpStats:
	case constant.OpStatsJob:
	case constant.OpStatsTube:
	case constant.OpListTubes:
	case constant.OpListTubeUsed:
	case constant.OpListTubesWatched:
	case constant.OpUse:
	case constant.OpWatch:
	case constant.OpIgnore:
	case constant.OpQuit:
	case constant.OpPauseTube:
	default:
		replyMsg(c, constant.MsgUnknownCommand)
	}
}

// displayOpPut 处理put指令
func displayOpPut(c *model.Coon) {
	idx := 4
	pri, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	delay, err := utils.ReadDuration(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	ttr, err := utils.ReadDuration(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	bodySize, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	utils.OpCt[constant.OpPut]++

	if bodySize > constant.JobDataSizeLimitDefault {
		skip(c, bodySize+2, constant.MsgJobTooBig)
		return
	}

	if c.Cmd[idx] != '\r' {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	// 设置连接类型成producer
	connSetProducer(c)

	if ttr < 1000000000 {
		ttr = 1000000000
	}

	// 创建job
	c.InJob = core.MakeJobWithID(uint32(pri), delay, ttr, bodySize+2, c.Use)

	// 填充body数据
	fillExtraData(c)
	maybeEnqueueInComingJob(c)
}

// displayOpPickReady
func displayOpPeekReady(c *model.Coon) {
	/* don't allow trailing garbage */
	if c.CmdLen != len(constant.CmdPeekReady)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	utils.OpCt[constant.OpPeekReady]++

	var j *model.Job
	if c.Use.Ready.Len() > 0 {
		r := c.Use.Ready.Take(0)
		if r != nil {
			j = core.JobCopy(r.Value.(*model.Job))
		}
	}

	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}
	replyJob(c, j, constant.MsgFound)
}

// displayOpPeekDelayed
func displayOpPeekDelayed(c *model.Coon) {
	if c.CmdLen != len(constant.CmdPeekDelayed)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpPeekDelayed]++

	var j *model.Job
	if c.Use.Delay.Len() > 0 {
		r := c.Use.Delay.Take(0)
		if r != nil {
			j = core.JobCopy(r.Value.(*model.Job))
		}
	}

	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}
	replyJob(c, j, constant.MsgFound)
}

// displayOpPeekBuried
func displayOpPeekBuried(c *model.Coon) {
	if c.CmdLen != len(constant.CmdPeekBuried)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpPeekBuried]++

	var j *model.Job
	if buriedJobP(c.Use) {
		j = core.JobCopy(c.Use.Buried.Next)
	}
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	replyJob(c, j, constant.MsgFound)
}

// displayOpPeekJob
func displayOpPeekJob(c *model.Coon) {
	idx := 0
	id, err := utils.ReadInt(c.Cmd[len(constant.CmdPeekJob):], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpPeekJob]++

	j := core.JobCopy(core.JobFind(uint64(id)))
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}
	replyJob(c, j, constant.MsgFound)
}

// displayOpReserveTimeout 中间状态用于从指令中获取timeout
func displayOpReserveTimeout(c *model.Coon) int64 {
	timeout, err := utils.StrTol(c.Cmd[len(constant.CmdReserveTimeout):])
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return -1
	}
	return timeout
}

// displayOpReserve 处理 reserve 指令
func displayOpReserve(c *model.Coon, timeout int64) {
	if c.CmdLen != len(constant.CmdReserve)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpReserve]++
	connSetWorker(c)

	if connDeadLineSoon(c) && !connReady(c) {
		replyMsg(c, constant.MsgDeadlineSoon)
		return
	}
	waitForJob(c, timeout)
	processQueue()
}

// displayOpReserveJob
func displayOpReserveJob(c *model.Coon) {
	idx := 0
	id, err := utils.ReadInt(c.Cmd[len(constant.CmdReserveJob):], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpReserveJob]++

	j := core.JobFind(uint64(id))
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	// 将job从相应状态的队列中删除
	switch j.R.State {
	case constant.Ready:
		j = removeReadyJob(j)
	case constant.Buried:
		j = removeBuriedJob(j)
	case constant.Delayed:
		j = removeDelayedJob(j)
	default:
		replyErr(c, constant.MsgInternalError)
		return
	}

	connSetWorker(c)
	utils.GlobalState.ReservedCt++

	connReserveJob(c, j)
	replyJob(c, j, constant.MsgReserved)
}

// displayOpDelete
func displayOpDelete(c *model.Coon) {
	idx := 0
	id, err := utils.ReadInt(c.Cmd[len(constant.CmdDelete):], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	utils.OpCt[constant.OpDelete]++

	var j *model.Job
	{
		jf := core.JobFind(uint64(id))
		j = removeReservedJob(c, jf)
		if j == nil {
			j = removeReadyJob(jf)
		}
		if j == nil {
			j = removeBuriedJob(jf)
		}
		if j == nil {
			j = removeDelayedJob(jf)
		}
	}

	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	j.Tube.Stat.TotalDeleteCt++
	j.R.State = constant.Invalid
	// TODO r = walwrite(&c->srv->wal, j);
	//  walmaint(&c->srv->wal);
	core.JobFree(j)
	// if (!r) {
	// 	reply_serr(c, MSG_INTERNAL_ERROR);
	// 	return;
	// }

	replyMsg(c, constant.MsgDeleted)
}

// displayOpRelease
func displayOpRelease(c *model.Coon) {
	idx := 0
	id, err := utils.ReadInt(c.Cmd[len(constant.CmdRelease):], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	pri, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	delay, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	utils.OpCt[constant.OpRelease]++

	j := removeReservedJob(c, core.JobFind(uint64(id)))
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	if delay > 0 {
		// 更新binlog
		// int z = walresvupdate(&c->srv->wal);
		// if (!z) {
		// 	reply_serr(c, MSG_OUT_OF_MEMORY);
		// 	return;
		// }
		// j->walresv += z;
	}

	j.R.Pri = uint32(pri)
	j.R.Delay = delay
	j.R.ReleaseCt++

	r := enqueueJob(c.Srv, j, delay, delay <= 0)
	if r < 0 {
		replyErr(c, constant.MsgInternalError)
		return
	}
	if r == 1 {
		replyMsg(c, constant.MsgReleased)
		return
	}

	core.BuryJob(c.Srv, j, false)
	replyMsg(c, constant.MsgBuried)
}

// displayOpBury
func displayOpBury(c *model.Coon) {
	idx := 0
	id, err := utils.ReadInt(c.Cmd[len(constant.CmdRelease):], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	pri, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpBury]++

	j := removeReservedJob(c, core.JobFind(uint64(id)))
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	j.R.Pri = uint32(pri)
	if !core.BuryJob(c.Srv, j, true) {
		replyErr(c, constant.MsgInternalError)
		return
	}
	replyMsg(c, constant.MsgBuried)
}

// displayOpKick
func displayOpKick(c *model.Coon) {
	count, err := utils.StrTol(c.Cmd[len(constant.CmdKick):])
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpKick]++

	i := kickJobs(c.Srv, c.Use, uint(count))
	replyLine(c, constant.StateSendWord, "KICKED %d\r\n", i)
}

// kickJobs
func kickJobs(s *model.Server, t *model.Tube, n uint) uint {
	if buriedJobP(t) {
		return kickBuriedJobs(s, t, n)
	}
	return kickDelayedJobs(s, t, n)
}

// kickBuriedJobs
func kickBuriedJobs(s *model.Server, t *model.Tube, n uint) uint {
	var i uint
	for i = 0; (i < n) && buriedJobP(t); i++ {
		kickBuriedJob(s, t.Buried.Next)
	}
	return i
}

// kickDelayedJobs
func kickDelayedJobs(s *model.Server, t *model.Tube, n uint) uint {
	var i uint
	for ; (i < n) && (t.Delay.Len() > 0); i++ {
		d := t.Delay.Take(0)
		if d == nil {
			continue
		}
		kickDelayedJob(s, d.Value.(*model.Job))
	}
	return i
}

// kickBuriedJob
func kickBuriedJob(s *model.Server, j *model.Job) bool {
	// TODO wal
	// z = walresvupdate(&s- > wal)
	// if !z
	// return 0
	// j- > walresv += z

	removeBuriedJob(j)

	j.R.KickCt++
	r := enqueueJob(s, j, 0, true)
	if r == 1 {
		return true
	}

	/* ready queue is full, so bury it */
	core.BuryJob(s, j, false)
	return false
}

// kickDelayedJob
func kickDelayedJob(s *model.Server, j *model.Job) bool {
	// TODO kickDelayedJob
	// z = walresvupdate(&s->wal);
	// if (!z)
	// return 0;
	// j->walresv += z;

	j.Tube.Delay.Remove(j.HeapIndex)
	j.R.KickCt++
	r := enqueueJob(s, j, 0, true)
	if r == 1 {
		return true
	}

	r = enqueueJob(s, j, j.R.Delay, false)
	if r == 1 {
		return false
	}

	core.BuryJob(s, j, false)
	return false
}

// fillExtraData 读取客户端额外data数据
func fillExtraData(c *model.Coon) {
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

func skip(c *model.Coon, n int64, msg []byte) {
	c.InJob = nil
	c.InJobRead = n
	fillExtraData(c)

	if c.InJobRead == 0 {
		reply(c, msg, int64(len(msg)), constant.StateSendWord)
		return
	}

	c.Reply = msg
	c.ReplyLen = int64(len(msg))
	c.ReplySent = 0
	c.State = constant.StateBitbucket
}

func connSetProducer(c *model.Coon) {
	if c.Type&constant.ConnTypeProducer > 0 {
		return
	}
	c.Type |= constant.ConnTypeProducer
	utils.CurProducerCt++
}

func connSetWorker(c *model.Coon) {
	if c.Type&constant.ConnTypeWorker > 0 {
		return
	}
	c.Type |= constant.ConnTypeWorker
	utils.CurWorkerCt++
}

func connSetSoonestJob(c *model.Coon, j *model.Job) {
	if c.SoonestJob == nil || j.R.DeadlineAt < c.SoonestJob.R.DeadlineAt {
		c.SoonestJob = j
	}
}

func connReserveJob(c *model.Coon, j *model.Job) {
	j.Tube.Stat.ReservedCt++
	j.R.ReserveCt++

	j.R.DeadlineAt = time.Now().UnixNano() + j.R.TTR
	j.R.State = constant.Reserved
	core.JobListInsert(c.ReservedJobs, j)
	j.Reservoir = c
	c.PendingTimeout = -1
	connSetSoonestJob(c, j)
}

func enqueueWaitingConn(c *model.Coon) {
	c.Type |= constant.ConnTypeWaiting
	utils.GlobalState.WaitingCt++
	c.Watch.Iterator(func(item interface{}) (bool, error) {
		t := item.(*model.Tube)
		t.Stat.WaitingCt++
		t.WaitingConns.Append(c)
		return false, nil
	})
}

func waitForJob(c *model.Coon, timeout int64) {
	c.State = constant.StateWait
	enqueueWaitingConn(c)
	/* Set the pending timeout to the requested timeout amount */
	c.PendingTimeout = int(timeout)
	// only care if they hang up
	EpollQAdd(c, 'h')
}

// remove_buried_job returns non-NULL value if job j was in the buried state.
// It excludes the job from the buried list and updates counters.
func removeBuriedJob(j *model.Job) *model.Job {
	if j == nil || j.R.State != constant.Buried {
		return nil
	}
	j = core.JobListRemove(j)
	if j != nil {
		utils.GlobalState.BuriedCt--
		j.Tube.Stat.BuriedCt--
	}
	return j
}

// removeReservedJob
func removeReservedJob(c *model.Coon, j *model.Job) *model.Job {
	if !isJobReservedByConn(c, j) {
		return nil
	}
	return removeThisReservedJob(c, j)
}

// isJobReservedByConn
func isJobReservedByConn(c *model.Coon, j *model.Job) bool {
	return j != nil && j.Reservoir == c && j.R.State == constant.Reserved
}

// remove_delayed_job returns non-NULL value if job j was in the delayed state.
// It removes the job from the tube delayed heap.
func removeDelayedJob(j *model.Job) *model.Job {
	if j == nil || j.R.State != constant.Delayed {
		return nil
	}
	j.Tube.Delay.Remove(j.HeapIndex)
	return j
}

// remove_ready_job returns non-NULL value if job j was in the ready state.
// It removes the job from the tube ready heap and updates counters.
func removeReadyJob(j *model.Job) *model.Job {
	if j == nil || j.R.State != constant.Ready {
		return nil
	}
	j.Tube.Ready.Remove(j.HeapIndex)
	utils.ReadyCt--
	if j.R.Pri < constant.UrgentThreshold {
		utils.GlobalState.UrgentCt--
		j.Tube.Stat.UrgentCt--
	}
	return j
}

// replyMsg
func replyMsg(c *model.Coon, m []byte) {
	reply(c, m, int64(len(m)), constant.StateSendWord)
}

// reply
func reply(c *model.Coon, line []byte, len int64, state int) {
	if c == nil {
		return
	}
	EpollQAdd(c, 'w')
	c.Reply = line
	c.ReplyLen = len
	c.ReplySent = 0
	c.State = state
}

// replyErr
func replyErr(c *model.Coon, err []byte) {
	fmt.Printf(" %s", err)
	replyMsg(c, append([]byte("server error: "), err...))
}

// replyLine
func replyLine(c *model.Coon, state int, fmtStr string, params ...interface{}) {
	c.ReplyBuf = []byte(fmt.Sprintf(fmtStr, params...))
	r := len(c.ReplyBuf)
	if r >= constant.LineBufSize {
		replyErr(c, constant.MsgInternalError)
	}
	reply(c, c.ReplyBuf, int64(r), state)
}

// replyJob
func replyJob(c *model.Coon, j *model.Job, msg []byte) {
	c.OutJob = j
	c.OutJobSent = 0
	replyLine(c, constant.StateSendJob, "%s %d %d\r\n",
		msg, j.R.ID, j.R.BodySize-2)
}

// processQueue
func processQueue() {
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
		c := ci.(*model.Coon)
		utils.GlobalState.ReservedCt++

		removeWaitingCoon(c)
		reserveJob(c, j)
		replyJob(c, j, constant.MsgReserved)
	}
}

// enqueueJob
func enqueueJob(s *model.Server, j *model.Job, delay int64, updateStore bool) int {
	j.Reservoir = nil
	if delay > 0 {
		j.R.DeadlineAt = time.Now().UnixNano() + delay
		if !j.Tube.Delay.Push(&structure.Item{Value: j}) {
			return 0
		}
		j.R.State = constant.Delayed
	} else {
		if !j.Tube.Ready.Push(&structure.Item{Value: j}) {
			return 0
		}
		j.R.State = constant.Ready
		utils.ReadyCt++
		if j.R.Pri < constant.UrgentThreshold {
			utils.GlobalState.UrgentCt++
			j.Tube.Stat.UrgentCt++
		}
	}

	if updateStore {
		// TODO 写入binlog
	}
	processQueue()
	return 1
}

// buriedJobP
func buriedJobP(t *model.Tube) bool {
	return !core.JobListEmpty(t.Buried)
}
