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

package ms_test

import (
	"fmt"
	"testing"

	"github.com/gostalk/gostalkd/internal/structure/ms"
)

type TestMs struct {
	A string
}

func TestNewMs(t *testing.T) {
	ms.NewMs()
}

func TestMs_NoLock(t *testing.T) {
	ms.NewMs().NoLock()
}

func TestMs_WithInsertEventFn(t *testing.T) {
	m := ms.NewMs().WithInsertFn(func(ms *ms.Ms, item interface{}, i int) {
		t.Logf("%#v, %#v, %d", ms, item, i)
	})
	m.Append(&TestMs{A: "a"})
}

func TestMs_WithDelEventFn(t *testing.T) {
	m := ms.NewMs().WithInsertFn(func(ms *ms.Ms, item interface{}, i int) {
		t.Logf("%#v, %#v, %d", ms, item, i)
	}).WithDelFn(func(ms *ms.Ms, item interface{}, i int) {
		fmt.Printf("%+v, %+v, %d\n", *ms, item, i)
	})
	m.Append(TestMs{A: "a"})
	m.Append(TestMs{A: "b"})
	m.DeleteWithIdx(0)
	t.Logf("%#v\n", m.Take())
}

func TestMs_Len(t *testing.T) {
	m := ms.NewMs().WithInsertFn(func(ms *ms.Ms, item interface{}, i int) {
		t.Logf("%#v, %#v, %d", ms, item, i)
	})
	m.Append(TestMs{
		A: "a",
	})
	t.Log(m.Len())
}

func TestMs_Iterator(t *testing.T) {
	m := ms.NewMs()
	t1 := TestMs{A: "a"}
	t2 := TestMs{A: "b"}
	m.Append(t1, t2)

	m.Iterator(func(item interface{}) bool {
		t.Log(item)
		return true
	})
	m.Iterator(func(item interface{}) bool {
		t.Log(item)
		return false
	})
}

func TestMs_Append(t *testing.T) {
	ms.NewMs().WithInsertFn(func(ms *ms.Ms, item interface{}, i int) {
		t.Logf("%#v, %#v, %d", ms, item, i)
	}).Append(TestMs{A: "a"})
}

func TestMs_Take(t *testing.T) {
	m := ms.NewMs()
	t1 := TestMs{A: "a"}
	t2 := TestMs{A: "b"}
	m.Append(t1, t2)

	t.Log(m.Take())
	t.Log(m.Take())
}

func TestMs_Contains(t *testing.T) {
	m := ms.NewMs()
	t1 := TestMs{
		A: "a",
	}
	t2 := &TestMs{
		A: "b",
	}
	m.Append(t1, t2)
	t.Log(m.Contains(t1))
	t.Log(m.Contains(t2))
	m.Contains(&TestMs{
		A: "c",
	})
}

func TestMs_DeleteWithIdx(t *testing.T) {
	m := ms.NewMs()
	t1 := TestMs{
		A: "a",
	}
	m.Append(t1)
	m.DeleteWithIdx(0)
	m.DeleteWithIdx(1)
	m.DeleteWithIdx(-1)
}

func TestMs_DeleteWithItem(t *testing.T) {
	m := ms.NewMs().WithDelFn(func(ms *ms.Ms, item interface{}, i int) {
		t.Logf("item: %#v idx: %d", item, i)
	})
	t1 := TestMs{
		A: "a",
	}
	m.Append(t1)
	m.DeleteWithItem(t1)
	m.DeleteWithItem(nil)
	m.DeleteWithItem(TestMs{
		A: "c",
	})
}

func TestMs_Clean(t *testing.T) {
	m := ms.NewMs()
	m.Append(TestMs{
		A: "a",
	})
	t.Log(m.Take())
	m.Clean()
	t.Log(m.Take())
}
