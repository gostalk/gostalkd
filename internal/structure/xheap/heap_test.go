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

package xheap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gostalk/gostalkd/internal/structure/xheap"
)

type TestHeapValue struct {
	A   string
	pri int64
	Idx int
}

func (t *TestHeapValue) Priority() int64 {
	return t.pri
}

func (t *TestHeapValue) SetIndex(idx int) {
	t.Idx = idx
}

func (t *TestHeapValue) Value() interface{} {
	return t.A
}

func (t *TestHeapValue) Index() int {
	return t.Idx
}

func TestNewItem(t *testing.T) {
	xheap.NewXHeap().Lock().Push(&TestHeapValue{})
}

func TestXHeap_NoLock(t *testing.T) {
	h := xheap.NewXHeap().NoLock()
	assert.Equal(t, h.LockModel(), false)
}

func TestXHeap_Lock(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	assert.Equal(t, h.LockModel(), true)
}

func TestItem_Index(t *testing.T) {
	item1 := &TestHeapValue{pri: 0}
	item2 := &TestHeapValue{pri: 1}
	xheap.NewXHeap().Lock().Push(item1, item2)
	assert.Equal(t, item2.Index(), 1)
}

func TestItem_Set(t *testing.T) {
	item := &TestHeapValue{A: "1"}
	xheap.NewXHeap().Lock().Push(item)

	assert.Equal(t, item.A, "1")
	item.A = "2"
	assert.Equal(t, item.A, "2")
}

func TestHeap_Push(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	h.Push(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	h.Push(nil)
}

func TestHeap_Pop(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	h.Push(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	v := h.Pop()
	t.Logf("value: %#v, index=%d", v, v.Index())

	assert.Nil(t, h.Pop())
}

func TestHeap_Take(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	v1 := &TestHeapValue{
		A:   "v1",
		pri: 1,
	}
	v2 := &TestHeapValue{
		A:   "v2",
		pri: 2,
	}
	h.Push(v1, v2)

	assert.Equal(t, h.Take(), v1)
	assert.Nil(t, h.Take(-1))
	assert.Nil(t, h.Take(2))
}

func TestXHeap_Fix(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	i1 := &TestHeapValue{
		A:   "v1",
		pri: 1,
	}
	i2 := &TestHeapValue{
		A:   "v2",
		pri: 2,
	}
	h.Push(i1, i2)

	assert.Equal(t, h.Take(), i1)

	i1.pri = 3

	h.Fix(0)
	h.Fix(-1)
	h.Fix(2)

	assert.Equal(t, h.Take(), i2)
}

func TestXHeap_RemoveWithIdx(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	v1 := &TestHeapValue{
		A:   "v1",
		pri: 1,
	}
	h.Push(v1)

	assert.Equal(t, h.RemoveByIdx(0), v1)
	assert.Nil(t, h.RemoveByIdx(0))
	assert.Nil(t, h.RemoveByIdx(-1))
}

func TestHeap_Len(t *testing.T) {
	h := xheap.NewXHeap().Lock()
	v1 := &TestHeapValue{
		A:   "v1",
		pri: 1,
	}
	h.Push(v1)

	assert.Equal(t, h.Len(), 1)
}
