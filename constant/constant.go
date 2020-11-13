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
	DefaultMaxJobSize            = 1<<16 - 1
	DefaultEachWriteAheadLogSize = 10485760
	DefaultPeriod                = int64(0x34630B8A000)
)

const (
	InstanceIDBytes     = 8
	JobDataSizeLimitMax = 1073741824
	MaxTubeNameLen      = 201 // The name of a tube cannot be longer than MaxTubeNameLen-1
	LineBufSize         = 11 + MaxTubeNameLen + 12
	AllJobsCap          = 12289
	UrgentThreshold     = 1024
	BucketBufSize       = 1024
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
	TotalOps

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

const (
	StatsFmt = "---\n" +
		"current-jobs-urgent: %d\n" +
		"current-jobs-ready: %d\n" +
		"current-jobs-reserved: %d\n" +
		"current-jobs-delayed: %d\n" +
		"current-jobs-buried: %d\n" +
		"cmd-put: %d\n" +
		"cmd-peek: %d\n" +
		"cmd-peek-ready: %d\n" +
		"cmd-peek-delayed: %d\n" +
		"cmd-peek-buried: %d\n" +
		"cmd-reserve: %d\n" +
		"cmd-reserve-with-timeout: %d\n" +
		"cmd-delete: %d\n" +
		"cmd-release: %d\n" +
		"cmd-use: %d\n" +
		"cmd-watch: %d\n" +
		"cmd-ignore: %d\n" +
		"cmd-bury: %d\n" +
		"cmd-kick: %d\n" +
		"cmd-touch: %d\n" +
		"cmd-stats: %d\n" +
		"cmd-stats-job: %d\n" +
		"cmd-stats-tube: %d\n" +
		"cmd-list-tubes: %d\n" +
		"cmd-list-tube-used: %d\n" +
		"cmd-list-tubes-watched: %d\n" +
		"cmd-pause-tube: %d\n" +
		"job-timeouts: %d\n" +
		"total-jobs: %d\n" +
		"max-job-size: %d\n" +
		"current-tubes: %d\n" +
		"current-connections: %d\n" +
		"current-producers: %d\n" +
		"current-workers: %d\n" +
		"current-waiting: %d\n" +
		"total-connections: %d\n" +
		"pid: %d\n" +
		"version: %s\n" +
		"rusage-utime: %d.%06d\n" +
		"rusage-stime: %d.%06d\n" +
		"uptime: %d\n" +
		"binlog-oldest-index: %d\n" +
		"binlog-current-index: %d\n" +
		"binlog-records-migrated: %d\n" +
		"binlog-records-written: %d\n" +
		"binlog-max-size: %d\n" +
		"draining: %s\n" +
		"id: %s\n" +
		"hostname: %s\n" +
		"os: %s\n" +
		"platform: %s\n" +
		"\r\n"

	StatsFmtTube = "---\n" +
		"name: %s\n" +
		"current-jobs-urgent: %d\n" +
		"current-jobs-ready: %d\n" +
		"current-jobs-reserved: %d\n" +
		"current-jobs-delayed: %d\n" +
		"current-jobs-buried: %d\n" +
		"total-jobs: %d\n" +
		"current-using: %d\n" +
		"current-watching: %d\n" +
		"current-waiting: %d\n" +
		"cmd-delete: %d\n" +
		"cmd-pause-tube: %d\n" +
		"pause: %d\n" +
		"pause-time-left: %d\n" +
		"\r\n"

	StatsFmtJob = "---\n" +
		"id: %d\n" +
		"tube: %s\n" +
		"state: %s\n" +
		"pri: %d\n" +
		"age: %d\n" +
		"delay: %d\n" +
		"ttr: %d\n" +
		"time-left: %d\n" +
		"file: %d\n" +
		"reserves: %d\n" +
		"timeouts: %d\n" +
		"releases: %d\n" +
		"buries: %d\n" +
		"kicks: %d\n" +
		"\r\n"
)
