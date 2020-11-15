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
	"time"

	"github.com/sjatsh/beanstalkd-go/constant"
	"github.com/sjatsh/beanstalkd-go/model"
	"github.com/sjatsh/beanstalkd-go/structure"
	"github.com/sjatsh/beanstalkd-go/utils"
)

var (
	nextJobId uint64 = 1
	allJobs          = make(map[uint64]*model.Job)
)

// NewJob 创建一个新Job对象
func NewJob(bodySize ...int64) *model.Job {
	j := &model.Job{
		R: model.JobRec{},
	}
	j.R.CreateAt = time.Now().UnixNano()
	if len(bodySize) > 0 {
		j.R.BodySize = bodySize[0]
	}
	ListReset(j)
	return j
}

// JobState 通过state id获取job详细状态
func JobState(state byte) string {
	switch state {
	case constant.Ready:
		return "ready"
	case constant.Reserved:
		return "reserved"
	case constant.Buried:
		return "buried"
	case constant.Delayed:
		return "delayed"
	default:
		return "invalid"
	}
}

// MakeJobWithID 创建一个带ID的job对象
func MakeJobWithID(pri uint32, delay, ttr, bodySize int64, tube *model.Tube, id ...uint64) *model.Job {
	j := NewJob(bodySize)
	if len(id) > 0 {
		j.R.ID = id[0]
		if j.R.ID >= nextJobId {
			nextJobId = j.R.ID + 1
		}
	} else {
		j.R.ID = nextJobId
		nextJobId++
	}

	j.R.Pri = pri
	j.R.Delay = delay
	j.R.TTR = ttr

	JobStore(j)

	j.Tube = tube
	return j
}

func BuryJob(s *model.Server, j *model.Job, updateStore bool) bool {
	if updateStore {
		z := WalResvUpdate(&s.Wal)
		if z <= 0 {
			return false
		}
		j.WalResv += z
	}

	JobListInsert(j, j.Tube.Buried)
	utils.GlobalState.BuriedCt++
	j.Tube.Stat.BuriedCt++
	j.R.State = constant.Buried
	j.Reservoir = nil
	j.R.BuryCt++

	if updateStore {
		if !WalWrite(&s.Wal, j) {
			return false
		}
		WalMain(&s.Wal)
	}

	return true
}

func ListReset(j *model.Job) {
	j.Prev = j
	j.Next = j
}

func JobStore(j *model.Job) {
	allJobs[j.R.ID] = j
	utils.AllJobsUsed++
}

func GetJobState(j *model.Job) string {
	state := JobState(j.R.State)
	return state
}

func JobFind(id uint64) *model.Job {
	j, ok := allJobs[id]
	if !ok {
		return nil
	}
	return j
}

func JobCopy(j *model.Job) *model.Job {
	if j == nil {
		return nil
	}
	j2 := *j
	n := &j2

	JobListRest(n)
	n.File = nil
	n.R.State = constant.Copy
	return n
}

func JobFree(j *model.Job) {
	if j != nil {
		j.Tube = nil
		if j.R.State != constant.Copy {
			delete(allJobs, j.R.ID)
		}
	}
}

func JobListRest(j *model.Job) {
	j.Prev = j
	j.Next = j
}

func JobListEmpty(j *model.Job) bool {
	isEmpty := j.Next == j && j.Prev == j
	return isEmpty
}

func JobListRemove(j *model.Job) *model.Job {
	if j == nil {
		return nil
	}
	if JobListEmpty(j) {
		return nil
	}
	j.Next.Prev = j.Prev
	j.Prev.Next = j.Next
	j.Prev = j
	j.Next = j
	return j
}

func JobListInsert(head, j *model.Job) {
	if !JobListEmpty(j) {
		return
	}
	j.Prev = head.Prev
	j.Next = head

	head.Prev.Next = j
	head.Prev = j
}

func JobSetPos(item *structure.Item, pos int) {
	item.Value.(*model.Job).HeapIndex = pos
}

func JobPriLess(ja, jb *structure.Item) bool {
	a := ja.Value.(*model.Job)
	b := jb.Value.(*model.Job)
	if a.R.Pri < b.R.Pri {
		return true
	}
	if a.R.Pri > b.R.Pri {
		return false
	}
	return a.R.ID < b.R.ID
}

func JobDelayLess(ja, jb *structure.Item) bool {
	a := ja.Value.(*model.Job)
	b := jb.Value.(*model.Job)
	if a.R.DeadlineAt < b.R.DeadlineAt {
		return true
	}
	if a.R.DeadlineAt > b.R.DeadlineAt {
		return false
	}
	return a.R.ID < b.R.ID
}
