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
package utils

import (
	"golang.org/x/sys/unix"

	"github.com/sjatsh/beanstalk-go/internal/constant"
)

var (
	// StartedAt serve started time
	StartedAt int64
	// OpCt Operation command count statistics
	OpCt = make([]constant.OpType, constant.OpTotal)
	// InstanceHex hex-encoded len of instance_id_bytes
	InstanceHex string
	// UtsName
	UtsName *unix.Utsname
	// AllJobsUsed
	AllJobsUsed int64

	ReadyCt   int64
	TimeoutCt uint64

	GlobalState = State{}
)

type State struct {
	UrgentCt      uint64
	WaitingCt     uint64
	BuriedCt      uint64
	ReservedCt    uint64
	PauseCt       uint64
	TotalDeleteCt uint64
	TotalJobsCt   uint64
}
