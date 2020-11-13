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
package core

import (
	"fmt"
	"sync/atomic"

	"github.com/sjatsh/beanstalk-go/constant"
	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/structure"
)

var (
	tubes       *structure.Ms // tubes链表
	defaultTube *model.Tube   // 默认tube
)

func init() {
	tubes = structure.NewMs()
	defaultTube = NewTube("default")
	tubes.Append(defaultTube)
}

// NewTube 创建一个新的tube
func NewTube(name string) *model.Tube {
	if len(name) > constant.MaxTubeNameLen {
		fmt.Println("truncating tube name")
	}
	t := &model.Tube{
		Name:  name,
		Ready: structure.NewHeap(),
		Delay: structure.NewHeap(),
	}
	// 设置排序规则方法
	t.Ready.SetLessFn(JobPriLess)
	t.Delay.SetLessFn(JobDelayLess)
	// 设置pos方法
	t.Ready.SetPosFn(JobSetPos)
	t.Delay.SetPosFn(JobSetPos)

	// 创建休眠队列
	t.Buried = NewJob()
	t.Buried.Prev = t.Buried
	t.Buried.Next = t.Buried

	t.WaitingConns = structure.NewMs()
	return t
}

func GetTubes() *structure.Ms {
	return tubes
}

func GetDefaultTube() *model.Tube {
	return defaultTube
}

func MakeAndInsertTube(name string) *model.Tube {
	t := NewTube(name)
	tubes.Append(t)
	return t
}

func SoonestDelayedJob() *model.Job {
	var j *model.Job
	tubes.Iterator(func(item interface{}) (bool, error) {
		t := item.(*model.Tube)
		if t.Delay.Len() == 0 {
			return false, nil
		}
		nji := t.Delay.Take()
		if nji == nil {
			return false, nil
		}
		nj := nji.Value.(*model.Job)
		if j == nil || nj.R.DeadlineAt < j.R.DeadlineAt {
			j = nj
		}
		return false, nil
	})
	return j
}

func NextAwaitedJob(now int64) *model.Job {
	var j *model.Job
	tubes.Iterator(func(item interface{}) (bool, error) {
		t := item.(*model.Tube)
		if t.Pause > 0 {
			if t.UnpauseAt > now {
				return false, nil
			}
			atomic.StoreInt64(&t.Pause, 0)
		}
		if t.WaitingConns.Len() > 0 && t.Ready.Len() > 0 {
			candidate := t.Ready.Take()
			if candidate == nil {
				return false, nil
			}
			if j == nil || JobPriLess(candidate, &structure.Item{Value: j}) {
				j = candidate.Value.(*model.Job)
			}
		}
		return false, nil
	})
	return j
}

func TubeFind(name string) *model.Tube {
	var t *model.Tube
	tubes.Iterator(func(item interface{}) (bool, error) {
		if item.(*model.Tube).Name == name {
			t = item.(*model.Tube)
			return true, nil
		}
		return false, nil
	})
	return t
}

func TubeFindOrMake(name string) *model.Tube {
	var t *model.Tube
	tubes.Iterator(func(item interface{}) (bool, error) {
		if item.(*model.Tube).Name == name {
			t = item.(*model.Tube)
			return true, nil
		}
		return false, nil
	})
	if t != nil {
		return t
	}
	return MakeAndInsertTube(name)
}
