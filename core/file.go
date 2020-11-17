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
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/gostalk/gostalkd/constant"
	"github.com/gostalk/gostalkd/model"
	"github.com/gostalk/gostalkd/utils"
)

var FAlloc = (FAllocH)(rawFAlloc)

type FAllocH func(*os.File, int64) bool

// rawFAlloc
func rawFAlloc(f *os.File, l int64) bool {
	var err error
	var w int
	buf := make([]byte, 4096)
	for i := 0; int64(i) < l; i += w {
		w, err = f.Write(buf)
		if err != nil {
			utils.Log.Errorln(err)
			return false
		}
	}
	f.Seek(0, 0)
	return true
}

// FileAddJob
func FileAddJob(f *model.File, j *model.Job) {
	h := &f.JobList
	if h.FPre == nil {
		h.FPre = h
	}
	j.File = f
	j.FPre = h.FPre
	j.FNext = h
	h.FPre.FNext = j
	h.FPre = j
	fileIncRef(f)
}

// FileRmJob
func FileRmJob(f *model.File, j *model.Job) {
	if f == nil {
		return
	}
	if f != j.File {
		return
	}
	j.FNext.FPre = j.FPre
	j.FPre.FNext = j.FNext

	j.FNext = nil
	j.FPre = nil
	j.File = nil

	f.W.Alive -= j.WalUsed
	j.WalUsed = 0
	fileDecRef(f)
}

// FileRead
func FileRead(f *model.File, list *model.Job) error {
	var version int64
	// 获取版本号
	if _, err := readFull(f, &version, "version"); err != nil {
		return err
	}

	var err error

	switch version {
	// v7版本
	case constant.WalVer:
		// 文件引用计数+1
		fileIncRef(f)
		// 循环读取文件中job信息
		for {
			if err = ReadRec(f, list); err != nil {
				break
			}
		}
		// 文件引用计数-1
		fileDecRef(f)
		return err
		// v5版本
	case constant.WalVer5:
		fileIncRef(f)
		for {
			if err = ReadRec5(f, list); err != nil {
				break
			}
		}
		fileDecRef(f)
		return err
	}

	unknown := fmt.Sprintf("%s: unknown version: %d\n", f.Path, version)
	utils.Log.Warn(unknown)
	return errors.New(unknown)
}

// ReadRec
func ReadRec(f *model.File, l *model.Job) error {
	var err error
	var size int64
	var nameLen int64
	var tubeName string

	if err := binary.Read(f.F, binary.LittleEndian, &nameLen); err != nil {
		utils.Log.Warnln("read name length error")
		warnPos(f, 0, err.Error())
		return err
	}
	size += int64(binary.Size(nameLen))

	if nameLen >= constant.MaxTubeNameLen {
		errStr := fmt.Sprintf("namelen %d exceeds maximum of %d", nameLen, constant.MaxTubeNameLen-1)
		warnPos(f, -binary.Size(nameLen), errStr)
		return errors.New(errStr)
	}

	if nameLen < 0 {
		warnPos(f, -binary.Size(nameLen), "name length %d is negative", nameLen)
		return err
	}

	if nameLen > 0 {
		nameBuf := make([]byte, nameLen)
		r, err := readFull(f, nameBuf, "tube name")
		if err != nil {
			return err
		}
		tubeName = string(nameBuf)
		size += int64(r)
	}

	jr := model.JobRec{}
	r, err := readFull(f, &jr, "job struct")
	if err != nil {
		return err
	}
	size += int64(r)

	if jr.ID <= 0 {
		return errors.New("job id is zero")
	}

	j := JobFind(jr.ID)
	if !(j != nil || nameLen > 0) {
		// We read a short record without having seen a
		// full record for this job, so the full record
		// was in an earlier file that has been deleted.
		// Therefore the job itself has either been
		// deleted or migrated; either way, this record
		// should be ignored.
		return nil
	}

	switch jr.State {
	case constant.Reserved:
		jr.State = constant.Ready
		fallthrough
	case constant.Ready, constant.Buried:
		fallthrough
	case constant.Delayed:
		if j == nil {
			if jr.BodySize > *utils.MaxJobSize {
				warnPos(f, -r, "job %d is too big (%d > %d)",
					jr.ID,
					jr.BodySize,
					utils.MaxJobSize)
				goto Error
			}

			t := TubeFind(tubeName)
			j = MakeJobWithID(jr.Pri, jr.Delay, jr.TTR, jr.BodySize, t, jr.ID)
			JobListRest(j)
			j.R.CreateAt = jr.CreateAt

			t.Stat.TotalJobsCt++
			utils.GlobalState.TotalJobsCt++
			if jr.State == constant.Buried {
				t.Stat.BuriedCt++
				utils.GlobalState.BuriedCt++
			}
		}

		j.R = jr
		JobListInsert(l, j)

		if nameLen > 0 {
			if jr.BodySize != j.R.BodySize {
				warnPos(f, -r, "job %d size changed", j.R.ID)
				warnPos(f, -r, "was %d, now %d", j.R.BodySize, jr.BodySize)
				goto Error
			}
			j.Body = make([]byte, j.R.BodySize)
			r, err := readFull(f, &j.Body, "job body")
			if err != nil {
				goto Error
			}
			size += int64(r)

			FileRmJob(j.File, j)
			FileAddJob(f, j)
		}
		j.WalUsed += size
		f.W.Alive += size
		return nil
	case constant.Invalid:
		if j != nil {
			JobListRemove(j)
			FileRmJob(j.File, j)
			JobFree(j)
		}
		return nil
	}
Error:
	if j != nil {
		JobListRemove(j)
		FileRmJob(j.File, j)
		JobFree(j)
	}
	return errors.New("")
}

// readRec5
func ReadRec5(f *model.File, l *model.Job) error {
	var err error
	var size int64
	var nameLen int64
	var tubeName string

	if err := binary.Read(f.F, binary.LittleEndian, &nameLen); err != nil {
		return err
	}
	size += int64(binary.Size(nameLen))

	if nameLen >= constant.MaxTubeNameLen {
		errStr := fmt.Sprintf("namelen %d exceeds maximum of %d", nameLen, constant.MaxTubeNameLen-1)
		warnPos(f, -binary.Size(nameLen), errStr)
		return errors.New(errStr)
	}
	if nameLen < 0 {
		warnPos(f, -binary.Size(nameLen), "name length %d is negative", nameLen)
		return err
	}
	if nameLen > 0 {
		r, err := readFull(f, &tubeName, "tube name")
		if err != nil {
			return err
		}
		size += int64(r)
	}

	jr := &model.JobRec5{}
	r, err := readFull(f, jr, "v5 job struct")
	if err != nil {
		return err
	}
	size += int64(r)

	if jr.ID <= 0 {
		return errors.New("job id is zero")
	}

	j := JobFind(jr.ID)

	if !(j != nil || nameLen > 0) {
		// We read a short record without having seen a
		// full record for this job, so the full record
		// was in an eariler file that has been deleted.
		// Therefore the job itself has either been
		// deleted or migrated; either way, this record
		// should be ignored.
		return nil
	}

	switch jr.State {
	case constant.Reserved:
		jr.State = constant.Ready
		fallthrough
	case constant.Ready:
		fallthrough
	case constant.Buried:
		fallthrough
	case constant.Delayed:
		if j == nil {
			if int64(jr.BodySize) > *utils.MaxJobSize {
				warnPos(f, -r, "job %d is too big (%d > %d)",
					jr.ID,
					jr.BodySize,
					utils.MaxJobSize)
				goto Error
			}
			t := TubeFindOrMake(tubeName)
			j = MakeJobWithID(jr.Pri, int64(jr.Delay), int64(jr.TTR), int64(jr.BodySize), t, jr.ID)
			JobListRest(j)
		}
		j.R.ID = jr.ID
		j.R.Pri = jr.Pri
		j.R.Delay = int64(jr.Delay * 1000)
		j.R.TTR = int64(jr.TTR * 1000)
		j.R.BodySize = int64(jr.BodySize)
		j.R.CreateAt = int64(jr.CreatedAt)
		j.R.DeadlineAt = int64(jr.DeadlineAt * 1000)
		j.R.ReserveCt = uint32(jr.ReserveCt)
		j.R.TimeoutCt = jr.TimeoutCt
		j.R.ReleaseCt = jr.ReleaseCt
		j.R.BuryCt = jr.BuryCt
		j.R.KickCt = jr.KickCt
		j.R.State = jr.State
		JobListInsert(l, j)

		if nameLen > 0 {
			if int64(jr.BodySize) != j.R.BodySize {
				warnPos(f, -r, "job %d size changed", j.R.ID)
				warnPos(f, -r, "was %d, now %d", j.R.BodySize, jr.BodySize)
				goto Error
			}
			r, err := readFull(f, j.Body, "v5 job body")
			if err != nil {
				goto Error
			}
			size += int64(r)
			// since this is a full record, we can move
			// the file pointer and decref the old
			// file, if any
			FileRmJob(j.File, j)
			FileAddJob(f, j)
		}
		j.WalUsed += size
		f.W.Alive += size
		return nil
	case constant.Invalid:
		if j != nil {
			JobListRemove(j)
			FileRmJob(j.File, j)
			JobFree(j)
		}
		return nil
	}
Error:
	if j != nil {
		JobListRemove(j)
		FileRmJob(j.File, j)
		JobFree(j)
	}
	return errors.New("")
}

// readFull
func readFull(f *model.File, i interface{}, desc string) (int, error) {
	if err := binary.Read(f.F, binary.LittleEndian, i); err != nil {
		warnPos(f, 0, "error reading %s", desc)
		return 0, err
	}
	return binary.Size(i), nil
}

// fileWOpen
func fileWOpen(f *model.File) {
	file, err := os.OpenFile(f.Path, os.O_WRONLY|os.O_CREATE, 0400)
	if err != nil {
		utils.Log.Warnf("open %s err:%s", f.Path, err)
		return
	}

	if !FAlloc(file, f.W.FileSize) {
		if err := file.Close(); err != nil {
			utils.Log.Warnf("unlink %s err:%s", f.Path, err)
		}
		if err := syscall.Unlink(f.Path); err != nil {
			utils.Log.Warnf("unlink %s err:%s", f.Path, err)
		}
		return
	}

	if err := binary.Write(file, binary.LittleEndian, constant.WalVer); err != nil {
		utils.Log.Panicf("write %s err:%s", f.Path, err)
	}

	f.F = file
	f.IsWOpen = true
	fileIncRef(f)
	f.Free = f.W.FileSize - int64(binary.Size(constant.WalVer))
	f.Resv = 0
}

func warnPos(f *model.File, adj int, fmtStr string, params ...interface{}) {
	off, err := f.F.Seek(0, os.SEEK_CUR)
	if err != nil {
		return
	}
	utils.Log.Errorf("%s:%d: %s", f.Path, off+int64(adj), fmt.Sprintf(fmtStr, params...))
}

func fileIncRef(f *model.File) {
	if f == nil {
		return
	}
	f.Refs++
}

func fileDecRef(f *model.File) {
	if f == nil {
		return
	}
	f.Refs--
	if f.Refs < 1 {
		walGc(f.W)
	}
}

// fileWrite
func fileWrite(f *model.File, j *model.Job, obj interface{}) bool {
	if err := binary.Write(f.F, binary.LittleEndian, obj); err != nil {
		return false
	}
	l := int64(binary.Size(obj))
	f.W.Resv -= l
	f.Resv -= l
	j.WalResv -= l
	j.WalUsed += l
	f.W.Alive += l
	return true
}

func fileWrJobShort(f *model.File, j *model.Job) bool {
	nl := int64(0)
	if !fileWrite(f, j, nl) {
		return false
	}
	if !fileWrite(f, j, j.R) {
		return false
	}

	if j.R.State == constant.Invalid {
		FileRmJob(j.File, j)
	}
	return true
}

func fileWrJobFull(f *model.File, j *model.Job) bool {
	FileAddJob(f, j)
	nl := int64(len(j.Tube.Name))
	return fileWrite(f, j, nl) &&
		fileWrite(f, j, []byte(j.Tube.Name)) &&
		fileWrite(f, j, j.R) &&
		fileWrite(f, j, j.Body)
}

func fileWClose(f *model.File) {
	if f == nil {
		return
	}
	if !f.IsWOpen {
		return
	}
	if f.Free > 0 {
		if err := f.F.Truncate(f.W.FileSize - f.Free); err != nil {
			utils.Log.Warnf("file Truncate %s", err)
		}
	}
	if err := f.F.Close(); err != nil {
		utils.Log.Warnf("file close %s", err)
	}
	f.IsWOpen = false
	fileDecRef(f)
}

func fileInit(f *model.File, w *model.Wal, n int64) bool {
	f.W = w
	f.Seq = n
	f.Path = fmt.Sprintf("%s/binlog.%d", w.Dir, n)
	return f.Path != ""
}

// Adds f to the linked list in w,
// updating w->tail and w->head as necessary.
func fileAdd(f *model.File, w *model.Wal) {
	if w.Tail != nil {
		w.Tail.Next = f
	}
	w.Tail = f
	if w.Head == nil {
		w.Head = f
	}
	w.NFile++
}
