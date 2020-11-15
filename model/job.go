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
	// persistent fields; these get written to the wal
	R JobRec

	// bookeeping fields; these are in-memory only
	Pad [6]byte

	Tube        *Tube
	Prev, Next  *Job
	HtNext      *Job
	HeapIndex   int
	File        *File
	FPre, FNext *Job
	Reservoir   *Coon

	WalResv int64
	WalUsed int64

	Body []byte
}
