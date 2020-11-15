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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sjatsh/beanstalk-go/model"
	"github.com/sjatsh/beanstalk-go/utils"
)

var Reserve = (ReserveH)(reserve)

type ReserveH func(w *model.Wal, n int64) int64

// WalInit
func WalInit(w *model.Wal, list *model.Job) {
	min := walScanDir(w)
	walRead(w, list, min)

	if !makeNextFile(w) {
		utils.Log.Panicln("make next file err")
	}
	w.Cur = w.Tail
}

// walScanDir
// Reads w->dir for files matching binlog.NNN,
// sets w->next to the next unused number, and
// returns the minimum number.
// If no files are found, sets w->next to 1 and
// returns a large number.
func walScanDir(w *model.Wal) int64 {
	var base = "binlog."
	var min, max int64 = 1 << 30, 0

	f, err := os.Open(w.Dir)
	if os.IsNotExist(err) {
		return min
	}
	if err := f.Close(); err != nil {
		utils.Log.Panic(err)
	}

	if err := filepath.Walk(w.Dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasPrefix(info.Name(), base) {
			return nil
		}
		idx, err := strconv.ParseInt(info.Name()[len(base):], 10, 64)
		if err != nil {
			return err
		}
		n := idx
		if n > max {
			max = n
		}
		if n < min {
			min = n
		}
		return nil
	}); err != nil {
		utils.Log.Panic(err)
	}

	w.Next = max + 1
	return min
}

func walGc(w *model.Wal) {
	var f *model.File
	for ; w.Head != nil && w.Head.Refs <= 0; {
		f = w.Head
		w.Head = f.Next
		if w.Tail == f {
			w.Tail = f.Next
		}

		w.NFile--
		if err := syscall.Unlink(f.Path); err != nil {
			utils.Log.Warnln(err)
			continue
		}
		f = nil
	}
}

func useNext(w *model.Wal) bool {
	f := w.Cur
	if f.Next == nil {
		utils.Log.Warnln("there is no next wal file")
		return false
	}
	w.Cur = f.Next
	fileWClose(f)
	return true
}

func ratio(w *model.Wal) int64 {
	var n, d int64
	d = w.Alive + w.Resv
	n = w.NFile*w.FileSize - d
	if d <= 0 {
		return 0
	}
	return n / d
}

func walResvMigrate(w *model.Wal, j *model.Job) int64 {
	z := int64(binary.Size(int64(1)))
	z += int64(len(j.Tube.Name))
	z += int64(binary.Size(j.R))
	z += j.R.BodySize
	return reserve(w, z)
}

func moveOne(w *model.Wal) {
	var j *model.Job
	if w.Head == w.Cur || w.Head.Next == w.Cur {
		return
	}

	j = w.Head.JobList.FNext
	if j == nil || j == &w.Head.JobList {
		utils.Log.Warnln("head holds no JobList")
		return
	}

	if walResvMigrate(w, j) <= 0 {
		return
	}

	FileRmJob(w.Head, j)
	w.Nmig++
	WalWrite(w, j)
}

func walCompact(w *model.Wal) {
	for r := ratio(w); r >= 2; r-- {
		moveOne(w)
	}
}

func WalSync(w *model.Wal) {
	now := time.Now().UnixNano()
	if !w.WantSync || now < w.LastSync+w.SyncRate {
		return
	}
	if err := w.Cur.F.Sync(); err != nil {
		utils.Log.Warnf("fsync error:%s\n", err)
	}
}

func WalWrite(w *model.Wal, j *model.Job) bool {
	ok := false
	if !w.Use {
		return true
	}
	if w.Cur.Resv > 0 || useNext(w) {
		if j.File != nil {
			ok = fileWrJobShort(w.Cur, j)
		} else {
			ok = fileWrJobFull(w.Cur, j)
		}
	}

	if !ok {
		fileWClose(w.Cur)
		w.Use = false
	}
	w.Nrec++
	return ok
}

func WalMain(w *model.Wal) {
	if w.Use {
		walCompact(w)
		WalSync(w)
	}
}

func makeNextFile(w *model.Wal) bool {
	f := &model.File{}
	if !fileInit(f, w, w.Next) {
		f = nil
		utils.Log.Warnln("init file err")
		return false
	}

	fileWOpen(f)
	if !f.IsWOpen {
		f = nil
		return false
	}

	w.Next++
	fileAdd(f, w)
	return true
}

func moveResv(to *model.File, from *model.File, n int64) {
	from.Resv -= n
	from.Free += n
	to.Resv += n
	to.Free -= n
}

func needFree(w *model.Wal, n int64) int64 {
	if w.Tail.Free >= n {
		return n
	}
	if makeNextFile(w) {
		return n
	}
	return 0
}

// Ensures:
//  1. b->resv is congruent to n (mod z).
//  2. x->resv is congruent to 0 (mod z) for each future file x.
// Assumes (and preserves) that b->resv >= n.
// Reserved space is conserved (neither created nor destroyed);
// we just move it around to preserve the invariant.
// We might have to allocate a new file.
// Returns 1 on success, otherwise 0. If there was a failure,
// w->tail is not updated.
func balanceRest(w *model.Wal, b *model.File, n int64) int64 {
	z := int64(binary.Size(int64(1)) + binary.Size(model.JobRec{}))
	if b == nil {
		return 1
	}

	rest := b.Resv - n
	r := rest % z
	if r == 0 {
		return balanceRest(w, b.Next, 0)
	}

	c := z - r
	if w.Tail.Resv >= c && b.Free >= c {
		moveResv(b, w.Tail, c)
		return balanceRest(w, b.Next, 0)
	}

	if needFree(w, r) != r {
		utils.Log.Warnln("need free")
		return 0
	}
	moveResv(w.Tail, b, r)
	return balanceRest(w, b.Next, 0)
}

// Ensures:
//  1. w->cur->resv >= n.
//  2. w->cur->resv is congruent to n (mod z).
//  3. x->resv is congruent to 0 (mod z) for each future file x.
// (where z is the size of a delete record in the wal).
// Reserved space is conserved (neither created nor destroyed);
// we just move it around to preserve the invariant.
// We might have to allocate a new file.
// Returns 1 on success, otherwise 0. If there was a failure,
// w->tail is not updated.
func balance(w *model.Wal, n int64) int64 {
	// Invariant 1
	// (this loop will run at most once)
	for ; w.Cur.Resv < n; {
		m := w.Cur.Resv
		if needFree(w, m) != m {
			utils.Log.Warnln("need free")
			return 0
		}

		moveResv(w.Tail, w.Cur, m)
		useNext(w)
	}
	return balanceRest(w, w.Cur, n)
}

func reserve(w *model.Wal, n int64) int64 {
	if !w.Use {
		return 1
	}
	if w.Cur.Free >= n {
		w.Cur.Free -= n
		w.Cur.Resv += n
		w.Resv += n
		return n
	}

	if needFree(w, n) != n {
		utils.Log.Warnln("need free")
		return 0
	}

	w.Tail.Free -= n
	w.Tail.Resv += n
	w.Resv += n

	if balance(w, n) == 0 {
		w.Resv -= n
		w.Tail.Resv -= n
		w.Tail.Free += n
		return 0
	}
	return n
}

// WalResvPut
func WalResvPut(w *model.Wal, j *model.Job) int64 {
	z := int64(binary.Size(int64(1)))
	z += int64(len(j.Tube.Name))
	z += int64(binary.Size(j.R))
	z += j.R.BodySize

	z += int64(binary.Size(int64(1)))
	z += int64(binary.Size(j.R))

	return reserve(w, z)
}

// WalResvUpdate
func WalResvUpdate(w *model.Wal) int64 {
	z := int64(binary.Size(int64(1)))
	z += int64(binary.Size(model.JobRec{}))
	return reserve(w, z)
}

// walDirLock
func WalDirLock(w *model.Wal) bool {
	if err := os.MkdirAll(w.Dir, os.ModePerm); err != nil {
		utils.Log.Errorf("mkdir err: %s", err)
		return false
	}

	path := w.Dir + "/lock"
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		utils.Log.Warnf("open %s err:%s", path, err)
		return false
	}

	lk := &syscall.Flock_t{
		Type:   syscall.F_WRLCK,
		Whence: 0,
		Start:  0,
		Len:    0,
	}
	if err := syscall.FcntlFlock(file.Fd(), syscall.F_SETLK, lk); err != nil {
		utils.Log.Warnf("fcntl err:%s", err)
		return false
	}
	return true
}

// walRead
func walRead(w *model.Wal, list *model.Job, min int64) {
	for i := min; i < w.Next; i++ {
		f := &model.File{}
		if !fileInit(f, w, i) {
			utils.Log.Panic("OOM")
		}

		file, err := os.OpenFile(f.Path, os.O_RDONLY, 0400)
		if err != nil {
			utils.Log.Error("open file err: %s", err)
			continue
		}

		f.F = file
		fileAdd(f, w)
		FileRead(f, list)
		file.Close()
	}
}
