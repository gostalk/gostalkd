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
	"sync/atomic"
	"time"

	"github.com/sjatsh/beanstalk-go/internal/structure"
	"github.com/sjatsh/beanstalk-go/internal/utils"
)

const (
	Invalid int32 = iota
	Ready
	Reserved
	Buried
	Delayed
	Copy
)

func JobState(state int32) string {
	switch state {
	case Ready:
		return "ready"
	case Reserved:
		return "reserved"
	case Buried:
		return "buried"
	case Delayed:
		return "delayed"
	default:
		return "invalid"
	}
}

var nextJobId uint64 = 1
var allJobs = make(map[uint64]*structure.Job)

func NewJob(bodySize ...int64) *structure.Job {
	j := &structure.Job{
		R: structure.JobRec{},
	}
	j.R.CreateAt = time.Now().UnixNano()
	if len(bodySize) > 0 {
		j.R.BodySize = bodySize[0]
	}
	ListReset(j)
	return j
}

func ListReset(j *structure.Job) {
	j.Prev = j
	j.Next = j
}

func JobStore(j *structure.Job) {
	allJobs[j.R.ID] = j
	utils.AllJobsUsed++
}

func GetJobState(j *structure.Job) string {
	state := JobState(j.R.State)
	return state
}

func JobFind(id uint64) *structure.Job {
	j, ok := allJobs[id]
	if !ok {
		return nil
	}
	return j
}

func JobCopy(j *structure.Job) *structure.Job {
	if j == nil {
		return nil
	}
	j2 := *j
	newJob := &j2
	newJob.File = nil
	newJob.R.State = Copy
	return newJob
}

func BuryJob(s *structure.Server, j *structure.Job, updateStore bool) bool {
	if updateStore {
		// TODO int z = walresvupdate(&s->wal);
		// if (!z)
		// return 0;
		// j->walresv += z;
	}

	JobListInsert(j, j.Tube.Buried)
	utils.GlobalState.BuriedCt++
	j.Tube.Stat.BuriedCt++
	j.R.State = Buried
	j.Reserver = nil
	j.R.BuryCt++

	if updateStore {
		// TODO if !walwrite(&s- > wal, j) {
		// 	return 0
		// }
		// walmaint(&s- > wal)
	}

	return true
}

func MakeJobWithID(pri uint32, delay, ttr, bodySize int64, tube *structure.Tube, id uint64)*structure.Job {
	j := NewJob(bodySize)
	if id > 0 {
		j.R.ID = id
		if id >= nextJobId {
			atomic.StoreUint64(&nextJobId, id+1)
		}
	} else {
		for {
			jobID := nextJobId
			if atomic.CompareAndSwapUint64(&nextJobId, jobID, jobID+1) {
				j.R.ID = jobID + 1
				break
			}
		}
	}
	j.R.Pri = pri
	j.R.Delay = delay
	j.R.TTR = ttr

	JobStore(j)

	j.Tube = tube
	return j
}

func JobFree(j *structure.Job) {
	if j != nil {
		j.Tube = nil
		if j.R.State != Copy {
			delete(allJobs, j.R.ID)
		}
	}
}

func JobListRest(j *structure.Job) {
	j.Prev = j
	j.Next = j
}

func JobListEmpty(j *structure.Job) bool {
	isEmpty := j.Next == j && j.Prev == j
	return isEmpty
}

func JobListRemove(j *structure.Job) *structure.Job {
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

func JobListInsert(head, j *structure.Job) {
	if !JobListEmpty(j) {
		return
	}
	j.Prev = j.Prev
	j.Next = j

	head.Prev.Next = j
	head.Prev = j
}

func JobSetPos(item *structure.Item, pos int) {
	item.Value.(*structure.Job).HeapIndex = pos
}

func JobPriLess(ja, jb *structure.Item) bool {
	a := ja.Value.(*structure.Job)
	b := jb.Value.(*structure.Job)
	if a.R.Pri < b.R.Pri {
		return true
	}
	if a.R.Pri > b.R.Pri {
		return false
	}
	return a.R.ID < b.R.ID
}

func JobDelayLess(ja, jb *structure.Item) bool {
	a := ja.Value.(*structure.Job)
	b := jb.Value.(*structure.Job)
	if a.R.DeadlineAt < b.R.DeadlineAt {
		return true
	}
	if a.R.DeadlineAt > b.R.DeadlineAt {
		return false
	}
	return a.R.ID < b.R.ID
}
