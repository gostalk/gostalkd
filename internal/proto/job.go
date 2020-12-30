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
	Invalid uint32 = iota
	Ready
	Reserved
	Buried
	Delayed
	Copy

	defaultJobSlot = 16384
)

type Jobs struct {
	jobs []*jobEntity
}

type jobEntity struct {
	sync.RWMutex
	m map[uint64]*Job
}

type Rec struct {
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
	State     uint32
}

type Job struct {
	R Rec

	heapIndex int

	body []byte

	tube *Tube

	// buried job
	listLock   sync.RWMutex
	Prev, Next *Job

	reservoir *Client
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

// FindByID
func (jobs *Jobs) FindByID(id uint64) *Job {
	jobsMap := jobs.jobs[id%defaultJobSlot]
	jobsMap.RLock()
	j := jobsMap.m[id]
	jobsMap.RUnlock()
	return j
}

// Store
func (jobs *Jobs) Store(j *Job) {
	jobsMap := jobs.jobs[j.R.ID%defaultJobSlot]
	jobsMap.Lock()
	jobsMap.m[j.R.ID] = j
	jobsMap.Unlock()
}

// Del
func (jobs *Jobs) Del(id uint64) {
	jobsMap := jobs.jobs[id%defaultJobSlot]
	jobsMap.Lock()
	delete(jobsMap.m, id)
	jobsMap.Unlock()
}

// NewJob
func NewJob(ids ...uint64) *Job {
	j := &Job{
		R: Rec{},
	}
	j.R.CreateAt = time.Now().UnixNano()
	var id uint64
	if len(ids) > 0 {
		id = ids[0]
		for {
			next := nextJobID
			if id > next {
				if atomic.CompareAndSwapUint64(&nextJobID, next, id+1) {
					break
				}
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
	allJobs.Store(j)
	return j
}

// Free
func (j *Job) Free() {
	j.tube = nil
	if j.R.State != Copy {
		allJobs.Del(j.R.ID)
	}
}

// ListReset
func (j *Job) ListReset() {
	j.Prev = j
	j.Next = j
}

// ListEmpty
func (j *Job) ListIsEmpty() bool {
	j.listLock.RLock()
	isEmpty := j.Next == j && j.Prev == j
	j.listLock.RUnlock()
	return isEmpty
}

// ListRemove
func (j *Job) ListRemove(removeJob *Job) *Job {
	if j.ListIsEmpty() {
		return nil
	}
	j.listLock.Lock()
	removeJob.Next.Prev = removeJob.Prev
	removeJob.Prev.Next = removeJob.Next
	removeJob.Prev = removeJob
	removeJob.Next = removeJob
	j.listLock.Unlock()
	return j
}

// ListInsert
func (j *Job) ListInsert(head, empty *Job) {
	if !j.ListIsEmpty() {
		return
	}
	j.listLock.Lock()
	empty.Prev = head.Prev
	empty.Next = head

	head.Prev.Next = empty
	head.Prev = empty
	j.listLock.Unlock()
}

// IsReserved
func (j *Job) IsReserved(c *Client) bool {
	return j != nil && j.reservoir == c && j.R.State == Reserved
}

// BuryJob
func (j *Job) BuryJob(buryJob *Job, updateStore bool) bool {
	// if updateStore {
	// 	z := WalResvUpdate(&s.Wal)
	// 	if z <= 0 {
	// 		return false
	// 	}
	// 	j.WalResv += z
	// }

	j.tube.buried.ListInsert(j, j.tube.buried)

	// TODO 统计
	atomic.StoreUint32(&j.R.State, Buried)
	j.reservoir = nil
	atomic.AddUint32(&j.R.BuryCt, 1)

	// if updateStore {
	// 	if !WalWrite(&s.Wal, j) {
	// 		return false
	// 	}
	// 	WalMain(&s.Wal)
	// }
	return true
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

func (j *Job) WithBodySize(size int64) *Job {
	j.R.BodySize = size
	return j
}

// Cover2DelayJob
func (j *Job) Cover2DelayJob() *DelayJob {
	return &DelayJob{Job: j}
}

// Cover2ReadyJob
func (j *Job) Cover2ReadyJob() *ReadyJob {
	return &ReadyJob{Job: j}
}

// Cover2ReservedJob
func (j *Job) Cover2ReservedJob() *ReservedJob {
	return &ReservedJob{Job: j}
}

type DelayJob struct {
	*Job
}

func (j *DelayJob) Priority() int64 {
	return j.R.DeadlineAt
}

func (j *DelayJob) SetIndex(idx int) {
	j.heapIndex = idx
}

func (j *DelayJob) Index() int {
	return j.heapIndex
}

type ReadyJob struct {
	*Job
}

func (j *ReadyJob) Priority() int64 {
	return int64(j.R.Pri)
}

func (j *ReadyJob) SetIndex(idx int) {
	j.heapIndex = idx
}

func (j *ReadyJob) Index() int {
	return j.heapIndex
}

type ReservedJob struct {
	*Job
}

func (j *ReservedJob) Priority() int64 {
	return j.R.DeadlineAt
}

func (j *ReservedJob) SetIndex(idx int) {
	j.heapIndex = idx
}

func (j *ReservedJob) Index() int {
	return j.heapIndex
}
