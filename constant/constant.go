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
package constant

// 默认server启动参数
const (
	DefaultPort                  = 11400
	DefaultFsyncMs               = 50
	DefaultListenAddr            = "0.0.0.0"
	DefaultMaxJobSize            = 65535
	DefaultEachWriteAheadLogSize = 10485760
	DefaultPeriod                = int64(0x34630B8A000)
)

const (
	InstanceIDBytes = 8
	// JobDataSizeLimitMax
	JobDataSizeLimitMax     = 1073741824
	JobDataSizeLimitDefault = 1<<16 - 1

	// MaxTubeNameLen
	MaxTubeNameLen = 201 // The name of a tube cannot be longer than MaxTubeNameLen-1
	LineBufSize    = 11 + MaxTubeNameLen + 12
	AllJobsCap     = 12289

	UrgentThreshold = 1024
	BucketBufSize   = 1024
)

// job相关状态
const (
	Invalid int32 = iota
	Ready
	Reserved
	Buried
	Delayed
	Copy
)

// 连接状态和指令相关常量
const (
	SafetyMargin = 1000000000

	ConnTypeProducer = 1 // 是生产者
	ConnTypeWorker   = 2 // 有reserve job添加WORKER类型
	ConnTypeWaiting  = 4 // 如果连接等待数据 设置WAITING类型

	StateWantCommand = iota // conn expects a command from the client
	StateWantData           // conn expects a job data
	StateSendJob            // conn sends job to the client
	StateSendWord           // conn sends a line reply
	StateWait               // client awaits for the job reservation
	StateBitbucket          // conn discards content
	StateClose              // conn should be closed
	StateWantEndLine        // skip until the end of a line

	OpUnknown = iota
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
	TpALOps

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
)

var (
	MsgNotFound       = []byte("NOT_FOUND\r\n")
	MsgFound          = []byte("FOUND")
	MsgReserved       = []byte("RESERVED")
	MsgDeadlineSoon   = []byte("DEADLINE_SOON\r\n")
	MsgTimedOut       = []byte("TIMED_OUT\r\n")
	MsgDeleted        = []byte("DELETED\r\n")
	MsgReleased       = []byte("RELEASED\r\n")
	MsgBuried         = []byte("BURIED\r\n")
	MsgKicked         = []byte("KICKED\r\n")
	MsgTouched        = []byte("TOUCHED\r\n")
	MsgBuriedFmt      = []byte("BURIED %d\r\n")
	MsgInsertedFmt    = []byte("INSERTED %d\r\n")
	MsgNotIgnored     = []byte("NOT_IGNORED\r\n")
	MsgOutOfMemory    = []byte("OUT_OF_MEMORY\r\n")
	MsgInternalError  = []byte("INTERNAL_ERROR\r\n")
	MsgDraining       = []byte("DRAINING\r\n")
	MsgBadFormat      = []byte("BAD_FORMAT\r\n")
	MsgUnknownCommand = []byte("UNKNOWN_COMMAND\r\n")
	MsgExpectedCrlf   = []byte("EXPECTED_CRLF\r\n")
	MsgJobTooBig      = []byte("JOB_TOO_BIG\r\n")

	// Cmd2OpMap
	Cmd2OpMap = map[string]int{
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
