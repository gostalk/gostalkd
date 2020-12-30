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
	"sync/atomic"
	"time"

	"github.com/gostalk/gostalkd/internal/network"
	"github.com/gostalk/gostalkd/internal/utils/str"
)

const (
	OpUnknown Option = iota
	OpPut
	OpPeekJob
	OpReserve
	OpDelete
	OpRelease
	OpBury
	OpKick
	OpStats
	OpStatsJob
	OpPeekBuried
	OpUse
	OpWatch
	OpIgnore
	OpListTubes
	OpListTubeUsed
	OpListTubesWatched
	OpStatsTube
	OpPeekReady
	OpPeekDelayed
	OpReserveTimeout
	OpTouch
	OpQuit
	OpPauseTube
	OpKickJob
	OpReserveJob
	TotalOps
)

const (
	CmdPut              = "put"
	CmdPeekJob          = "peek"
	CmdPeekReady        = "peek-ready"
	CmdPeekDelayed      = "peek-delayed"
	CmdPeekBuried       = "peek-buried"
	CmdReserve          = "reserve"
	CmdReserveTimeout   = "reserve-with-timeout"
	CmdReserveJob       = "reserve-job"
	CmdDelete           = "delete"
	CmdRelease          = "release"
	CmdBury             = "bury"
	CmdKick             = "kick"
	CmdKickJob          = "kick-job"
	CmdTouch            = "touch"
	CmdStats            = "stats"
	CmdStatsJob         = "stats-job"
	CmdUse              = "use"
	CmdWatch            = "watch"
	CmdIgnore           = "ignore"
	CmdListTubes        = "list-tubes"
	CmdListTubeUsed     = "list-tube-used"
	CmdListTubesWatched = "list-tubes-watched"
	CmdStatsTube        = "stats-tube"
	CmdQuit             = "quit"
	CmdPauseTube        = "pause-tube"
)

const (
	MsgNotFound       = "NOT_FOUND"
	MsgFound          = "FOUND"
	MsgReserved       = "RESERVED"
	MsgDeadlineSoon   = "DEADLINE_SOON"
	MsgTimedOut       = "TIMED_OUT"
	MsgDeleted        = "DELETED"
	MsgReleased       = "RELEASED"
	MsgBuried         = "BURIED"
	MsgKicked         = "KICKED"
	MsgTouched        = "TOUCHED"
	MsgBuriedFmt      = "BURIED %d"
	MsgInsertedFmt    = "INSERTED %d"
	MsgNotIgnored     = "NOT_IGNORED"
	MsgOutOfMemory    = "OUT_OF_MEMORY"
	MsgInternalError  = "INTERNAL_ERROR"
	MsgDraining       = "DRAINING"
	MsgBadFormat      = "BAD_FORMAT"
	MsgUnknownCommand = "UNKNOWN_COMMAND"
	MsgExpectedCrlf   = "EXPECTED_CRLF"
	MsgJobTooBig      = "JOB_TOO_BIG"
)

var (
	// Cmd2Option
	Cmd2Option = map[string]Option{
		CmdPut:              OpPut,
		CmdPeekJob:          OpPeekJob,
		CmdPeekReady:        OpPeekReady,
		CmdPeekDelayed:      OpPeekDelayed,
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
)

type (
	Option int

	ProcessEntity struct {
		ProcessState CliState
		Session      network.Session
		Reply        network.Reply
		InJob        *Job
		Use          string
	}
)

// ioProcess
func ioProcess(client *Client) error {
	switch client.state {
	case StateWantCommand:
		return dispatchCmd(client)
	case StateWantData:
		l := int64(len(client.GetSession().Data()))
		if client.inJob.R.BodySize != l {
			client.ReplyMsg(MsgBadFormat)
			return nil
		}
		client.inJob.body = client.session.Data()

		enqueueJob(client)

		client.inJob = nil
		client.state = StateWantCommand
	}
	return nil
}

// dispatchCmd
func dispatchCmd(client *Client) error {
	cmd := string(client.session.Data())
	tokens := str.Parse(cmd)

	var cmdName string
	if err := tokens.NextString(&cmdName).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return err
	}

	option, ok := Cmd2Option[cmdName]
	if !ok {
		client.ReplyMsg(MsgNotFound)
		return nil
	}

	switch option {
	case OpPut:
		dispatchPut(client, tokens)
	case OpPeekReady:
		dispatchPeekReady(client)
	case OpPeekDelayed:
		dispatchPeekDelayed(client)
	case OpPeekBuried:
		dispatchPeekBuried(client)
	}

	return nil
}

// dispatchPut
func dispatchPut(client *Client, tokens *str.Tokens) {
	if tokens.Len() < 5 {
		client.reply.WriteString(MsgBadFormat)
		return
	}

	var (
		pri                  uint32
		delay, ttr, bodySize int64
	)
	if err := tokens.NextUint32(&pri).
		NextInt64(&delay).
		NextInt64(&ttr).
		NextInt64(&bodySize).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return
	}

	if ttr < int64(time.Second) {
		ttr = int64(time.Second)
	}

	j := NewJob().WithPri(pri).
		WithDelay(delay).
		WithTTR(ttr).
		WithBodySize(bodySize).
		WithTube(client.use)

	client.inJob = j
	client.state = StateWantData
}

// dispatchPeekReady
func dispatchPeekReady(client *Client) {
	var ready *ReadyJob
	if client.use.readyJobs.Len() > 0 {
		j := client.use.readyJobs.Take()
		if j != nil {
			ready = j.(*ReadyJob)
		}
	}

	if ready == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}
	client.ReplyJob(ready.Job, MsgFound)
}

// dispatchPeekDelayed
func dispatchPeekDelayed(client *Client) {
	var delayed *DelayJob
	if client.use.delayJobs.Len() > 0 {
		j := client.use.delayJobs.Take()
		if j != nil {
			delayed = j.(*DelayJob)
		}
	}

	if delayed == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}
	client.ReplyJob(delayed.Job, MsgFound)
}

// dispatchPeekBuried
func dispatchPeekBuried(client *Client) {
	var j interface{}
	if client.use.buried > 0 {
		j = client.use.buried.Take()
	}
	if j == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}
	client.ReplyJob(j.(*Job), MsgFound)
}

// dispatchOpPeekJob
func dispatchOpPeekJob(client *Client, tokens *str.Tokens) {
	var id uint64
	if err := tokens.NextUint64(&id).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return
	}

	j := allJobs.FindByID(id)
	if j == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}
	client.ReplyJob(j, MsgFound)
}

// dispatchOpReserveTimeout
func dispatchOpReserveTimeout(client *Client, tokens *str.Tokens) int64 {
	var timeout int64
	if err := tokens.NextInt64(&timeout).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return -1
	}
	return timeout
}

// dispatchOpReserve
func dispatchOpReserve(client *Client, t Option, timeout int64) {
	if connDeadLineSoon(client) && !client.WatchTubeHaveReady() {
		client.ReplyMsg(MsgDeadlineSoon)
		return
	}
	client.WaitForJob(timeout)
}

// dispatchOpReserveJob
func dispatchOpReserveJob(client *Client, tokens *str.Tokens) {
	var id uint64
	if err := tokens.NextUint64(&id).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return
	}
	j := allJobs.FindByID(id)
	if j == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}
	switch j.R.State {
	case Ready:
		j = j.tube.RemoveReadyJob(j)
	case Delayed:
		j = j.tube.RemoveDelayedJob(j)
	case Buried:
		j = j.tube.RemoveBuriedJob(j)
	default:
		client.ReplyErr(MsgInternalError)
		return
	}

	client.SetWorker()
	client.ReplyJob(j, MsgReserved)
}

func dispatchOpDelete(client *Client, tokens *str.Tokens) {
	var id uint64
	if err := tokens.NextUint64(&id).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return
	}
	jf := allJobs.FindByID(id)
	if jf == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}

	var j *Job
	{
		j = removeReservedJob(client, jf)
		if j == nil {
			j = jf.tube.RemoveReadyJob(j)
		}
		if j == nil {
			j = jf.tube.RemoveBuriedJob(jf)
		}
		if j == nil {
			j = jf.tube.RemoveDelayedJob(j)
		}
	}

	if j == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}

	atomic.StoreUint32(&j.R.State, Invalid)

	// TODO 统计
	// TODO 删除binlog

	j.Free()

	client.ReplyMsg(MsgDeleted)
}

// removeReservedJob
func removeReservedJob(client *Client, j *Job) *Job {
	if j == nil || !j.IsReserved(client) {
		return nil
	}
	return j.ListRemove(j)
}

// dispatchOpRelease
func dispatchOpRelease(client *Client, tokens *str.Tokens) {
	var (
		id    uint64
		pri   uint32
		delay int64
	)

	if err := tokens.NextUint64(&id).
		NextUint32(&pri).
		NextInt64(&delay).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return
	}

	// TODO 统计

	j := removeReservedJob(client, allJobs.FindByID(id))
	if j == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}

	// TODO 写入binlog
	if delay > 0 {

	}

	j.R.Pri = pri
	j.R.Delay = delay
	atomic.AddUint32(&j.R.ReleaseCt, 1)

	enqueueJob(client)
	client.ReplyMsg(MsgReleased)
}

// dispatchOpBury
func dispatchOpBury(client *Client, tokens *str.Tokens) {
	var (
		id  uint64
		pri uint32
	)
	if err := tokens.NextUint64(&id).
		NextUint32(&pri).Err(); err != nil {
		client.ReplyErr(MsgBadFormat)
		return
	}

	j := removeReservedJob(client, allJobs.FindByID(id))
	if j == nil {
		client.ReplyMsg(MsgNotFound)
		return
	}

	j.R.Pri = pri
	if !j.BuryJob(j, true) {
		client.ReplyErr(MsgInternalError)
		return
	}

	client.ReplyMsg(MsgBuried)
}

func dispatchOpKick(client *Client, tokens *str.Tokens) {
	var count int
	if err := tokens.NextInt(&count).Err(); err != nil {
		client.ReplyMsg(MsgBadFormat)
		return
	}

	t := client.use

	var i int
	if !client.use.buried.ListIsEmpty() {
		for ; i < count && !t.buried.ListIsEmpty(); i++ {
			kickBuriedJob(client, client.use.buried.Next)
		}
	} else {
		for ; i < count && t.delayJobs.Len() > 0; i++ {
			d := t.delayJobs.Take(0)
			if d == nil {
				continue
			}

		}
	}

}

func kickBuriedJob(client *Client, j *Job) {
	// z := core.WalResvUpdate(&s.Wal)
	// if z <= 0 {
	// 	return false
	// }
	// j.WalResv += z
	j.tube.RemoveBuriedJob(j)

	atomic.AddUint32(&j.R.KickCt, 1)
	enqueueJob(client)
}

// kickDelayedJob
func kickDelayedJob(j *Job) {
	// z := core.WalResvUpdate(&s.Wal)
	// if z <= 0 {
	// 	return false
	// }
	// j.WalResv += z

	j.tube.RemoveDelayedJob(j)
	atomic.AddUint32(&j.R.KickCt, 1)

}

// connDeadLineSoon
func connDeadLineSoon(client *Client) bool {
	j := client.reservedJobs.Take()
	if j == nil {
		return false
	}
	if time.Now().UnixNano() < j.(*ReservedJob).R.DeadlineAt-int64(time.Second) {
		return false
	}
	return true
}

// enqueueJob
func enqueueJob(client *Client) {
	if client.inJob.R.Delay > 0 {
		client.inJob.R.State = Delayed
		client.inJob.R.DeadlineAt = time.Now().UnixNano() + client.inJob.R.Delay
		client.use.PushDelayJob(client.inJob.Cover2DelayJob())
	} else {
		client.inJob.R.State = Ready
		client.use.PushReadyJob(client.inJob.Cover2ReadyJob())
	}
}
