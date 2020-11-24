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

type (
	CliState int32
	Option   int
)

const (
	StateWantCommand CliState = iota // client expects a command from the client
	StateWantData                    // client expects a job data
	StateBitbucket                   // client discards content
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
