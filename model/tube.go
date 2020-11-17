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
package model

import (
	"github.com/gostalk/gostalkd/structure"
)

type Tube struct {
	Name         string
	Ready        *structure.Heap
	Delay        *structure.Heap
	WaitingConns *structure.Ms
	Stat         State // job各个状态统计
	UsingCt      int   // tube正在被多少coon监听
	WatchingCt   int   // tube上watching的coon

	Pause     int64 // 暂停时间，单位nsec，pause-tube 命令设置
	UnpauseAt int64 // 暂停结束时间的时间戳

	Buried *Job // 休眠状态job链表
}

type State struct {
	UrgentCt      uint64
	WaitingCt     uint64
	BuriedCt      uint64
	ReservedCt    uint64
	PauseCt       uint64
	TotalDeleteCt uint64
	TotalJobsCt   uint64
}
