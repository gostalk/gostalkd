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
	pri int
}

func TestNewItem(t *testing.T) {
	xheap.NewItem(1)
}

func TestXHeap_NoLock(t *testing.T) {
	xheap.NewXHeap().NoLock()
}

func TestItem_Index(t *testing.T) {
	assert.Equal(t, xheap.NewItem(1).Index(), 0)
}

func TestItem_Value(t *testing.T) {
	assert.Equal(t, xheap.NewItem(1).Value(), 1)
}

func TestItem_Set(t *testing.T) {
	item := xheap.NewItem(1)
	assert.Equal(t, item.Value(), 1)

	item.Set(2)
	assert.Equal(t, item.Value(), 2)
}

func TestHeap_Push(t *testing.T) {
	h := xheap.NewXHeap()
	h.SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	h.Push(xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	}))
	h.Push(nil)
}

func TestHeap_Pop(t *testing.T) {
	h := xheap.NewXHeap()
	h.SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	h.Push(xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	}))
	v1 := h.Pop()
	t.Logf("value: %#v, index=%d", v1.Value(), v1.Index())

	assert.Nil(t, h.Pop())
}

func TestHeap_Take(t *testing.T) {
	h := xheap.NewXHeap()
	h.SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	v1 := xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	v2 := xheap.NewItem(&TestHeapValue{
		A:   "v2",
		pri: 2,
	})

	h.Push(v1, v2)
	assert.Equal(t, h.Take(), v1)
	assert.Nil(t, h.Take(-1))
	assert.Nil(t, h.Take(2))
}

func TestXHeap_Fix(t *testing.T) {
	h := xheap.NewXHeap().SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	i1 := xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	i2 := xheap.NewItem(&TestHeapValue{
		A:   "v2",
		pri: 2,
	})
	h.Push(i1, i2)

	assert.Equal(t, h.Take(), i1)

	i1.Value().(*TestHeapValue).pri = 3
	h.Fix(0)
	h.Fix(-1)
	h.Fix(2)

	assert.Equal(t, h.Take(), i2)
}

func TestXHeap_RemoveWithIdx(t *testing.T) {
	h := xheap.NewXHeap()
	h.SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	v1 := xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	h.Push(v1)

	assert.Equal(t, h.RemoveWithIdx(0), v1)
	assert.Nil(t, h.RemoveWithIdx(0))
	assert.Nil(t, h.RemoveWithIdx(-1))
}

func TestXHeap_RemoveWithItem(t *testing.T) {
	h := xheap.NewXHeap()
	h.SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	v1 := xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	h.Push(v1)

	assert.Equal(t, h.RemoveWithItem(v1), v1)
	assert.Nil(t, h.RemoveWithItem(v1))
	assert.Nil(t, h.RemoveWithItem(nil))
}

func TestHeap_Len(t *testing.T) {
	h := xheap.NewXHeap()
	h.SetLessFn(func(i, j *xheap.Item) bool {
		return i.Value().(*TestHeapValue).pri < j.Value().(*TestHeapValue).pri
	})

	v1 := xheap.NewItem(&TestHeapValue{
		A:   "v1",
		pri: 1,
	})
	h.Push(v1)

	assert.Equal(t, h.Len(), 1)
}
