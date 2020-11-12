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

	"github.com/sjatsh/beanstalk-go/internal/constant"
	"github.com/sjatsh/beanstalk-go/internal/structure"
)

var (
	tubes       = structure.NewMs()         // tubes链表
	defaultTube = TubeFindOrMake("default") // 默认tube
)

func NewTube(name string) *structure.Tube {
	if len(name) > constant.MaxTubeNameLen {
		fmt.Println("truncating tube name")
	}
	t := &structure.Tube{
		Name:  name,
		Ready: structure.NewHeap(),
		Delay: structure.NewHeap(),
	}
	t.Ready.SetLessFn(JobPriLess)
	t.Delay.SetLessFn(JobDelayLess)
	t.Ready.SetPosFn(JobSetPos)
	t.Delay.SetPosFn(JobSetPos)

	j := NewJob()
	t.Buried = j
	t.Buried.Prev = t.Buried
	t.Buried.Next = t.Buried

	t.WaitingConns = structure.NewMs()
	return t
}

func GetTubes() *structure.Ms {
	return tubes
}

func GetDefaultTube() *structure.Tube {
	return defaultTube
}

func MakeAndInsertTube(name string) *structure.Tube {
	t := NewTube(name)
	tubes.Append(t)
	return t
}

func SoonestDelayedJob() *structure.Job {
	var j *structure.Job
	tubes.Iterator(func(item interface{}) (bool, error) {
		t := item.(*structure.Tube)
		if t.Delay.Len() == 0 {
			return false, nil
		}
		nji := t.Delay.Take()
		if nji == nil {
			return false, nil
		}
		nj := nji.Value.(*structure.Job)
		if j == nil || nj.R.DeadlineAt < j.R.DeadlineAt {
			j = nj
		}
		return false, nil
	})
	return j
}

func NextAwaitedJob(now int64) *structure.Job {
	var j *structure.Job
	tubes.Iterator(func(item interface{}) (bool, error) {
		t := item.(*structure.Tube)
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
				j = candidate.Value.(*structure.Job)
			}
		}
		return false, nil
	})
	return j
}

func TubeFind(name string) *structure.Tube {
	var t *structure.Tube
	tubes.Iterator(func(item interface{}) (bool, error) {
		if item.(*structure.Tube).Name == name {
			t = item.(*structure.Tube)
			return true, nil
		}
		return false, nil
	})
	return t
}

func TubeFindOrMake(name string) *structure.Tube {
	var t *structure.Tube
	tubes.Iterator(func(item interface{}) (bool, error) {
		if item.(*structure.Tube).Name == name {
			t = item.(*structure.Tube)
			return true, nil
		}
		return false, nil
	})
	if t != nil {
		return t
	}
	return MakeAndInsertTube(name)
}
