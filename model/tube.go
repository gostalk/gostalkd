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
	"github.com/sjatsh/beanstalk-go/structure"
	"github.com/sjatsh/beanstalk-go/utils"
)

type Tube struct {
	Name         string
	Ready        *structure.Heap
	Delay        *structure.Heap
	WaitingConns *structure.Ms
	Stat         utils.State
	// struct stats stat;              // job各个状态统计
	UsingCt    int // tube被多少coon监听
	WatchingCt int // waiting连接个数,coon加入waiting链表的时候+1 删除时-1

	Pause     int64 // 暂停时间，单位nsec，pause-tube 命令设置
	UnpauseAt int64 // 暂停结束时间的时间戳

	Buried *Job // 休眠状态job链表
}
