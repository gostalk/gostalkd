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
	"math"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gostalk/gostalkd/constant"
	"github.com/gostalk/gostalkd/core"
	"github.com/gostalk/gostalkd/model"
	"github.com/gostalk/gostalkd/structure"
	"github.com/gostalk/gostalkd/utils"
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
	for epollQ != nil {
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
	for epollQ != nil {
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

	defer func() {
		if err := recover(); err != nil {
			utils.Log.Errorf("prot tick panic error:%s\n", err)
		}
	}()

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
		item := s.Connes.Take()
		c := item.Value.(*model.Coon)
		d = c.TickAt - now
		if d > 0 {
			period = int64(math.Min(float64(period), float64(d)))
			break
		}
		s.Connes.Remove(item.Index())
		c.InCoons = 0
		connTimeout(c)
	}
	EpollQApply()
	return time.Duration(period)
}

// whichCmd 验证命令是否有效
func whichCmd(cmdBuf []byte) int {
	cmdBuf = cmdBuf[:utils.StrLen(cmdBuf)]
	cmdSlice := bytes.Split(cmdBuf, []byte(" "))
	cmd := string(cmdSlice[0])

	if op, ok := constant.Cmd2OpMap[cmd]; ok {
		return op
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

	t := whichCmd(c.Cmd)

	switch t {
	case constant.OpPut:
		dispatchOpPut(c)
	case constant.OpPeekReady:
		dispatchOpPeekReady(c)
	case constant.OpPeekDelayed:
		dispatchOpPeekDelayed(c)
	case constant.OpPeekBuried:
		dispatchOpPeekBuried(c)
	case constant.OpPeekJob:
		dispatchOpPeekJob(c)
	case constant.OpReserveTimeout:
		timeout = dispatchOpReserveTimeout(c)
		fallthrough
	case constant.OpReserve:
		dispatchOpReserve(c, t, timeout)
	case constant.OpReserveJob:
		dispatchOpReserveJob(c)
	case constant.OpDelete:
		dispatchOpDelete(c)
	case constant.OpRelease:
		dispatchOpRelease(c)
	case constant.OpBury:
		dispatchOpBury(c)
	case constant.OpKick:
		dispatchOpKick(c)
	case constant.OpKickJob:
		dispatchOpKickJob(c)
	case constant.OpTouch:
		dispatchOpTouch(c)
	case constant.OpStats:
		dispatchStats(c)
	case constant.OpStatsJob:
		dispatchStatsJob(c)
	case constant.OpStatsTube:
		dispatchStatsTube(c)
	case constant.OpListTubes:
		dispatchOpListTubes(c)
	case constant.OpListTubeUsed:
		dispatchOpListTubeUsed(c)
	case constant.OpListTubesWatched:
		dispatchOpListTubesWatched(c)
	case constant.OpUse:
		dispatchOpUse(c)
	case constant.OpWatch:
		dispatchOpWatch(c)
	case constant.OpIgnore:
		dispatchOpIgnore(c)
	case constant.OpQuit:
		dispatchOpQuit(c)
	case constant.OpPauseTube:
		dispatchOpPauseTube(c)
	default:
		replyMsg(c, constant.MsgUnknownCommand)
	}
}

// dispatchOpPut 处理put指令
func dispatchOpPut(c *model.Coon) {
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

	if bodySize > *utils.MaxJobSize {
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

// dispatchOpPeekReady
func dispatchOpPeekReady(c *model.Coon) {
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

// dispatchOpPeekDelayed
func dispatchOpPeekDelayed(c *model.Coon) {
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

// dispatchOpPeekBuried
func dispatchOpPeekBuried(c *model.Coon) {
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

// dispatchOpPeekJob
func dispatchOpPeekJob(c *model.Coon) {
	idx := len(constant.CmdPeekJob)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
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

// dispatchOpReserveTimeout 中间状态用于从指令中获取timeout
func dispatchOpReserveTimeout(c *model.Coon) int64 {
	timeout, err := utils.StrTol(c.Cmd[len(constant.CmdReserveTimeout)+1:])
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return -1
	}
	return timeout
}

// dispatchOpReserve 处理 reserve 指令
func dispatchOpReserve(c *model.Coon, t int, timeout int64) {
	if t == constant.OpReserve && c.CmdLen != len(constant.CmdReserve)+2 {
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

// dispatchOpReserveJob
func dispatchOpReserveJob(c *model.Coon) {
	idx := len(constant.CmdReserveJob)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
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

// dispatchOpDelete
func dispatchOpDelete(c *model.Coon) {
	idx := len(constant.CmdDelete)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
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

	j.Tube.Stat.TotalJobsCt--
	j.Tube.Stat.TotalDeleteCt++
	utils.GlobalState.TotalJobsCt--
	utils.GlobalState.TotalDeleteCt++

	j.R.State = constant.Invalid
	r := core.WalWrite(&c.Srv.Wal, j)
	core.WalMain(&c.Srv.Wal)
	core.JobFree(j)
	if !r {
		replyErr(c, constant.MsgInternalError)
		return
	}
	replyMsg(c, constant.MsgDeleted)
}

// dispatchOpRelease
func dispatchOpRelease(c *model.Coon) {
	idx := len(constant.CmdRelease)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
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
		z := core.WalResvUpdate(&c.Srv.Wal)
		if z <= 0 {
			replyErr(c, constant.MsgOutOfMemory)
			return
		}
		j.WalResv += z
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

// dispatchOpBury
func dispatchOpBury(c *model.Coon) {
	idx := len(constant.CmdRelease)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
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

// dispatchOpKick
func dispatchOpKick(c *model.Coon) {
	count, err := utils.StrTol(c.Cmd[len(constant.CmdKick):])
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpKick]++

	i := kickJobs(c.Srv, c.Use, uint(count))
	replyLine(c, constant.StateSendWord, "KICKED %d\r\n", i)
}

// dispatchOpKickJob
func dispatchOpKickJob(c *model.Coon) {
	idx := len(constant.CmdRelease)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	utils.OpCt[constant.OpKickJob]++

	j := core.JobFind(uint64(id))
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	if !((j.R.State == constant.Buried && kickBuriedJob(c.Srv, j)) ||
		(j.R.State == constant.Delayed && kickDelayedJob(c.Srv, j))) {
		replyMsg(c, constant.MsgNotFound)
		return
	}
	replyMsg(c, constant.MsgKicked)
}

// dispatchOpTouch
func dispatchOpTouch(c *model.Coon) {
	idx := len(constant.CmdRelease)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpTouch]++

	if !touchJob(c, core.JobFind(uint64(id))) {
		replyMsg(c, constant.MsgNotFound)
		return
	}
	replyMsg(c, constant.MsgTouched)
}

// dispatchStats
func dispatchStats(c *model.Coon) {
	if c.CmdLen != len(constant.CmdStats)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpStats]++
	doStats(c, constant.OpStats)
}

// dispatchStatsJob
func dispatchStatsJob(c *model.Coon) {
	idx := len(constant.CmdStatsJob)
	id, err := utils.ReadInt(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpStatsJob]++

	j := core.JobFind(uint64(id))
	if j == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	if j.Tube == nil {
		replyErr(c, constant.MsgInternalError)
		return
	}
	doStats(c, constant.OpStatsJob, j)
}

// dispatchStatsTube
func dispatchStatsTube(c *model.Coon) {
	idx := len(constant.CmdStatsTube)
	name, err := utils.ReadTubeName(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	if !utils.IsValidTube(name) {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpStatsTube]++

	t := core.TubeFind(string(name))
	if t == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}
	doStats(c, constant.OpStatsTube, t)
}

// dispatchOpListTubes
func dispatchOpListTubes(c *model.Coon) {
	if c.CmdLen != len(constant.CmdListTubes)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpListTubes]++
	doListTubes(c, core.GetTubes())
}

// dispatchOpListTubeUsed
func dispatchOpListTubeUsed(c *model.Coon) {
	if c.CmdLen != len(constant.CmdListTubeUsed)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpListTubeUsed]++
	replyLine(c, constant.StateSendWord, "USING %s\r\n", c.Use.Name)
}

// dispatchOpListTubesWatched
func dispatchOpListTubesWatched(c *model.Coon) {
	if c.CmdLen != len(constant.CmdListTubesWatched)+2 {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpListTubesWatched]++
	doListTubes(c, c.Watch)
}

// dispatchOpUse
func dispatchOpUse(c *model.Coon) {
	idx := len(constant.CmdUse)
	name, err := utils.ReadTubeName(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	if !utils.IsValidTube(name) {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpUse]++

	t := core.TubeFindOrMake(string(name))
	c.Use.UsingCt--
	c.Use = t
	c.Use.UsingCt++

	replyLine(c, constant.StateSendWord, "USING %s\r\n", c.Use.Name)
}

// dispatchOpWatch
func dispatchOpWatch(c *model.Coon) {
	idx := len(constant.CmdWatch)
	name, err := utils.ReadTubeName(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	if !utils.IsValidTube(name) {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpWatch]++

	t := core.TubeFindOrMake(string(name))
	if !c.Watch.Contains(t) {
		c.Watch.Append(t)
	}
	replyLine(c, constant.StateSendWord, "WATCHING %d\r\n", c.Watch.Len())
}

// dispatchOpIgnore
func dispatchOpIgnore(c *model.Coon) {
	idx := len(constant.CmdIgnore)
	name, err := utils.ReadTubeName(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	if !utils.IsValidTube(name) {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpIgnore]++

	var t *model.Tube
	c.Watch.Iterator(func(item interface{}) (bool, error) {
		if item.(*model.Tube).Name == string(name) {
			t = item.(*model.Tube)
			return true, nil
		}
		return false, nil
	})

	if t != nil && c.Watch.Len() < 2 {
		replyMsg(c, constant.MsgNotIgnored)
		return
	}

	if t != nil {
		c.Watch.Remove(t)
	}
	t = nil
	replyLine(c, constant.StateSendWord, "WATCHING %d\r\n", c.Watch.Len())
}

// dispatchOpQuit
func dispatchOpQuit(c *model.Coon) {
	c.State = constant.StateClose
}

// dispatchOpPauseTube
func dispatchOpPauseTube(c *model.Coon) {
	idx := len(constant.CmdPauseTube)
	name, err := utils.ReadTubeName(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	delay, err := utils.ReadDuration(c.Cmd[idx:], &idx)
	if err != nil {
		replyMsg(c, constant.MsgBadFormat)
		return
	}
	utils.OpCt[constant.OpPauseTube]++

	if !utils.IsValidTube(name) {
		replyMsg(c, constant.MsgBadFormat)
		return
	}

	t := core.TubeFind(string(name))
	if t == nil {
		replyMsg(c, constant.MsgNotFound)
		return
	}

	if delay == 0 {
		delay = 1
	}

	t.UnpauseAt = time.Now().UnixNano() + delay
	t.Pause = delay
	t.Stat.PauseCt++

	replyLine(c, constant.StateSendWord, "PAUSED\r\n")
}

// doStats
func doStats(c *model.Coon, t int, data ...interface{}) {

	var reply string
	switch t {
	case constant.OpStats:
		reply = fmtStats(c.Srv)
	case constant.OpStatsJob:
		reply = fmtStatsJob(data[0].(*model.Job))
	case constant.OpStatsTube:
		reply = fmtStatsTube(data[0].(*model.Tube))
	default:
		return
	}

	replyLen := int64(len(reply))

	c.OutJob = core.NewJob(replyLen)
	if c.OutJob == nil {
		replyErr(c, constant.MsgOutOfMemory)
		return
	}
	/* Mark this job as a copy so it can be appropriately freed later on */
	c.OutJob.R.State = constant.Copy

	/* now actually format the stats data */
	c.OutJob.Body = []byte(reply)
	c.OutJobSent = 0
	replyLine(c, constant.StateSendJob, "OK %d\r\n", replyLen-2)
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
	z := core.WalResvUpdate(&s.Wal)
	if z <= 0 {
		return false
	}
	j.WalResv += z

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
	z := core.WalResvUpdate(&s.Wal)
	if z <= 0 {
		return false
	}
	j.WalResv += z

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

// skip
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

// connSetProducer
func connSetProducer(c *model.Coon) {
	if c.Type&constant.ConnTypeProducer > 0 {
		return
	}
	c.Type |= constant.ConnTypeProducer
	utils.CurProducerCt++
}

// connSetWorker
func connSetWorker(c *model.Coon) {
	if c.Type&constant.ConnTypeWorker > 0 {
		return
	}
	c.Type |= constant.ConnTypeWorker
	utils.CurWorkerCt++
}

// connSetSoonestJob
func connSetSoonestJob(c *model.Coon, j *model.Job) {
	if c.SoonestJob == nil || j.R.DeadlineAt < c.SoonestJob.R.DeadlineAt {
		c.SoonestJob = j
	}
}

// connReserveJob
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

// enqueueWaitingConn
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

// waitForJob
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

// touchJob
func touchJob(c *model.Coon, j *model.Job) bool {
	if !isJobReservedByConn(c, j) {
		return false
	}
	j.R.DeadlineAt = time.Now().UnixNano() + j.R.TTR
	c.SoonestJob = nil
	return true
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

	l := string(line)
	if l == string(constant.MsgBadFormat) || l == string(constant.MsgInternalError) ||
		l == string(constant.MsgNotFound) || l == string(constant.MsgUnknownCommand) ||
		l == string(constant.MsgTimedOut) {
		log := utils.Log.WithFields(map[string]interface{}{
			"cmd":       string(c.Cmd),
			"cmd_len":   c.CmdLen,
			"cmd_read":  c.CmdRead,
			"reply":     string(line),
			"reply_len": len,
			"state":     state,
		})
		if c.Use != nil {
			log = log.WithFields(map[string]interface{}{
				"use": map[string]interface{}{
					"name":        c.Use.Name,
					"stat":        c.Use.Stat,
					"pause":       c.Use.Pause,
					"unpause_at":  c.Use.UnpauseAt,
					"using_ct":    c.Use.UsingCt,
					"watching_ct": c.Use.WatchingCt,
				},
			})
		}
		if c.InJob != nil {
			log = log.WithFields(map[string]interface{}{
				"in_job":      c.InJob.R,
				"in_job_body": string(c.InJob.Body),
				"in_job_read": c.InJobRead,
			})
		}
		if c.OutJob != nil {
			log = log.WithFields(map[string]interface{}{
				"out_job":      c.OutJob.R,
				"out_job_body": string(c.OutJob.Body),
				"out_job_sent": c.OutJobSent,
			})
		}
		log.Warnln(l)
	}

	EpollQAdd(c, 'w')
	c.Reply = line
	c.ReplyLen = len
	c.ReplySent = 0
	c.State = state
}

// replyErr
func replyErr(c *model.Coon, err []byte) {
	replyErr := fmt.Sprintf("server error: %s", string(err))
	utils.Log.Error(replyErr)
	replyMsg(c, []byte(replyErr))
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
		utils.ReadyCt--
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
		// job写入binlog
		if !core.WalWrite(&s.Wal, j) {
			return 0
		}
		// 文件拆分&文件fsync
		core.WalMain(&s.Wal)
	}
	processQueue()
	return 1
}

// buriedJobP
func buriedJobP(t *model.Tube) bool {
	return !core.JobListEmpty(t.Buried)
}

// getDelayedJobCt
func getDelayedJobCt() int {
	var count int
	core.GetTubes().Iterator(func(item interface{}) (bool, error) {
		count += item.(*model.Tube).Delay.Len()
		return false, nil
	})
	return count
}

func uptime() int64 {
	return (time.Now().UnixNano() - utils.StartedAt) / 1000000000
}

// fmtStats
func fmtStats(s *model.Server) string {
	var whead, wcur int64

	if s.Wal.Head != nil {
		whead = s.Wal.Head.Seq
	}

	if s.Wal.Cur != nil {
		wcur = s.Wal.Cur.Seq
	}

	// utsName := utils.GetUtsName()
	ru := utils.Getrusage()
	info := utils.GetSysInfo()

	drainMode := strconv.FormatBool(atomic.LoadInt64(&utils.DrainMode) == 1)

	return fmt.Sprintf(constant.StatsFmt,
		utils.GlobalState.UrgentCt,
		utils.ReadyCt,
		utils.GlobalState.ReservedCt,
		getDelayedJobCt(),
		utils.GlobalState.BuriedCt,
		utils.OpCt[constant.OpPut],
		utils.OpCt[constant.OpPeekJob],
		utils.OpCt[constant.OpPeekReady],
		utils.OpCt[constant.OpPeekDelayed],
		utils.OpCt[constant.OpPeekBuried],
		utils.OpCt[constant.OpReserve],
		utils.OpCt[constant.OpReserveTimeout],
		utils.OpCt[constant.OpDelete],
		utils.OpCt[constant.OpRelease],
		utils.OpCt[constant.OpUse],
		utils.OpCt[constant.OpWatch],
		utils.OpCt[constant.OpIgnore],
		utils.OpCt[constant.OpBury],
		utils.OpCt[constant.OpKick],
		utils.OpCt[constant.OpTouch],
		utils.OpCt[constant.OpStats],
		utils.OpCt[constant.OpStatsJob],
		utils.OpCt[constant.OpStatsTube],
		utils.OpCt[constant.OpListTubes],
		utils.OpCt[constant.OpListTubeUsed],
		utils.OpCt[constant.OpListTubesWatched],
		utils.OpCt[constant.OpPauseTube],
		utils.TimeoutCt,
		utils.GlobalState.TotalJobsCt,
		utils.MaxJobSize,
		core.GetTubes().Len(),
		utils.CurConnCt,
		utils.CurProducerCt,
		utils.CurWorkerCt,
		utils.GlobalState.WaitingCt,
		utils.TotalConnCt,
		os.Getpid(),
		utils.Version,
		ru.Usec,
		ru.Total,
		ru.Ssec,
		ru.Total,
		uptime(),
		whead,
		wcur,
		s.Wal.Nmig,
		s.Wal.Nrec,
		s.Wal.FileSize,
		drainMode,
		utils.InstanceHex,
		info.NodeName,
		info.Version,
		info.Machine,
	)
}

// fmtStatsJob
func fmtStatsJob(j *model.Job) string {
	var file int64
	var timeLeft int64
	t := time.Now().UnixNano()

	if j.R.State == constant.Reserved || j.R.State == constant.Delayed {
		timeLeft = (j.R.DeadlineAt - t) / 1000000000
	}
	if j.File != nil {
		file = j.File.Seq
	}
	return fmt.Sprintf(constant.StatsFmtJob,
		j.R.ID,
		j.Tube.Name,
		core.JobState(j.R.State),
		j.R.Pri,
		(t-j.R.CreateAt)/1000000000,
		j.R.Delay/1000000000,
		j.R.TTR/1000000000,
		timeLeft,
		file,
		j.R.ReserveCt,
		j.R.TimeoutCt,
		j.R.ReleaseCt,
		j.R.BuryCt,
		j.R.KickCt,
	)
}

// fmtStatsTube
func fmtStatsTube(t *model.Tube) string {
	var timeLeft int64
	if t.Pause > 0 {
		timeLeft = (t.UnpauseAt - time.Now().UnixNano()) / 1000000000
	}
	return fmt.Sprintf(constant.StatsFmtTube,
		t.Name,
		t.Stat.UrgentCt,
		t.Ready.Len(),
		t.Stat.ReservedCt,
		t.Delay.Len(),
		t.Stat.BuriedCt,
		t.Stat.TotalJobsCt,
		t.UsingCt,
		t.WatchingCt,
		t.Stat.WaitingCt,
		t.Stat.TotalDeleteCt,
		t.Stat.PauseCt,
		t.Pause/1000000000,
		timeLeft,
	)
}

// doListTubes
func doListTubes(c *model.Coon, l *structure.Ms) {
	var respZ int64 = 6
	var t *model.Tube

	l.Iterator(func(item interface{}) (bool, error) {
		t = item.(*model.Tube)
		respZ += 3 + int64(len(t.Name))
		return false, nil
	})

	c.OutJob = core.NewJob(respZ)
	c.OutJob.R.State = constant.Copy
	c.OutJob.Body = append(c.OutJob.Body, "---\n"...)

	l.Iterator(func(item interface{}) (bool, error) {
		t = item.(*model.Tube)
		c.OutJob.Body = append(c.OutJob.Body, "- "+t.Name+"\n"...)
		return false, nil
	})
	c.OutJob.Body = append(c.OutJob.Body, "\r\n"...)
	c.OutJobSent = 0
	replyLine(c, constant.StateSendJob, "OK %d\r\n", respZ-2)
}

// protReplay
func protReplay(s *model.Server, list *model.Job) bool {

	var nj *model.Job
	for j := list.Next; j != list; j = nj {
		nj = j.Next
		core.JobListRemove(j)
		z := core.WalResvUpdate(&s.Wal)
		if z <= 0 {
			utils.Log.Warnln("failed to reserve space")
			return false
		}

		var delay int64
		switch j.R.State {
		case constant.Buried:
			core.BuryJob(s, j, false)
		case constant.Delayed:
			t := time.Now().UnixNano()
			if t < j.R.DeadlineAt {
				delay = j.R.DeadlineAt - t
			}
			fallthrough
		default:
			if enqueueJob(s, j, delay, false) < 1 {
				utils.Log.Warnf("error recovering job %d", j.R.ID)
			}
		}
	}
	return true
}
