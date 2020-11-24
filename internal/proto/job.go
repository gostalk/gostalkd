// Copyright 2020 The Gostalkd Project Authors.
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

package proto

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	nextJobID uint64 = 1
	allJobs   *Jobs
)

const (
	JobInvalid byte = iota
	JobReady
	JobReserved
	JobBuried
	JobDelayed
	JobCopy

	defaultJobSlot = 1024
)

type Jobs struct {
	jobs []*jobEntity
}

type jobEntity struct {
	sync.RWMutex
	m map[uint64]*Job
}

type JobRec struct {
	ID       uint64
	Pri      uint32
	Delay    int64
	TTR      int64
	BodySize int64
	CreateAt int64

	// deadline_at is a timestamp, in nsec, that points to:
	// * time when job will become ready for delayed job,
	// * time when TTR is about to expire for reserved job,
	// * undefined otherwise.
	DeadlineAt int64

	ReserveCt uint32
	TimeoutCt uint32
	ReleaseCt uint32
	BuryCt    uint32
	KickCt    uint32
	State     byte
}

type Job struct {
	R    JobRec
	tube *Tube

	heapIndex int

	body []byte
}

// NewJobs
func init() {
	jobs := &Jobs{
		jobs: make([]*jobEntity, defaultJobSlot),
	}
	for idx := range jobs.jobs {
		jobs.jobs[idx] = &jobEntity{
			m: make(map[uint64]*Job),
		}
	}
	allJobs = jobs
}

// AllJobs
func AllJobs() *Jobs {
	return allJobs
}

// Put
func (jobs *Jobs) Put(j *Job) {
	jobsMap := jobs.jobs[j.R.ID%defaultJobSlot]
	jobsMap.Lock()
	jobsMap.m[j.R.ID] = j
	jobsMap.Unlock()
}

// NewJob
func NewJob(ids ...uint64) *Job {
	j := &Job{
		R: JobRec{},
	}
	j.R.CreateAt = time.Now().UnixNano()
	var id uint64
	if len(ids) > 0 {
		id = ids[0]
		for {
			next := nextJobID
			if id > next {
				atomic.CompareAndSwapUint64(&nextJobID, next, id+1)
			}
		}
	} else {
		for {
			id = nextJobID
			if atomic.CompareAndSwapUint64(&nextJobID, id, id+1) {
				break
			}
		}
	}

	j.R.ID = id
	allJobs.Put(j)
	return j
}

func (j *Job) WithPri(pri uint32) *Job {
	j.R.Pri = pri
	return j
}

func (j *Job) WithDelay(delay int64) *Job {
	j.R.Delay = delay
	return j
}

func (j *Job) WithTTR(ttr int64) *Job {
	j.R.TTR = ttr
	return j
}

func (j *Job) WithTube(tube *Tube) *Job {
	j.tube = tube
	return j
}

func (j *Job) GetTube() *Tube {
	return j.tube
}

func (j *Job) WithBodySize(size int64) *Job {
	j.R.BodySize = size
	return j
}

// Cover2DelayJob
func (j *Job) Cover2DelayJob() *DelayJob {
	return (*DelayJob)(j)
}

// Cover2ReadyJob
func (j *Job) Cover2ReadyJob() *ReadyJob {
	return (*ReadyJob)(j)
}

// Body
func (j *Job) Body() []byte {
	return j.body
}

type DelayJob Job

func (j *DelayJob) Priority() int64 {
	return j.R.DeadlineAt
}

func (j *DelayJob) SetIndex(idx int) {
	j.heapIndex = idx
}

func (j *DelayJob) Index() int {
	return j.heapIndex
}

type ReadyJob Job

func (j *ReadyJob) Priority() int64 {
	return int64(j.R.Pri)
}

func (j *ReadyJob) SetIndex(idx int) {
	j.heapIndex = idx
}

func (j *ReadyJob) Index() int {
	return j.heapIndex
}
