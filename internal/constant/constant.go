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
	JobDataSizeLimitMax = 1073741824
	// MaxTubeNameLen
	MaxTubeNameLen = 201 // The name of a tube cannot be longer than MaxTubeNameLen-1
	LineBufSize    = 11 + MaxTubeNameLen + 12
	AllJobsCap     = 12289

	UrgentThreshold = 1024
)

// OpType
type OpType int

const (
	OpUnknown OpType = iota
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
	OpListTubesUsed
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
	OpTotal
)
