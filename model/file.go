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
	"os"
	"unsafe"
)

type JobRec5 struct {
	ID         uint64
	Pri        uint32
	Delay      uint64
	TTR        uint64
	BodySize   int32
	CreatedAt  uint64
	DeadlineAt uint64
	ReserveCt  uint64
	TimeoutCt  uint32
	ReleaseCt  uint32
	BuryCt     uint32
	KickCt     uint32
	State      byte
	Pad        [1]byte
}

type File struct {
	Next    *File
	Refs    uint
	Seq     int
	IsWOpen int
	F       *os.File
	Free    int
	Resv    int
	Path    string

	W       *Wal
	JobList Job
}

var JobRec5Size = int64(unsafe.Offsetof(JobRec5{}.Pad))
