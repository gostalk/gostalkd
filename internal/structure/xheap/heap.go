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

package xheap

import (
	"container/heap"
	"sync"
)

// XHeap
type XHeap struct {
	l    sync.RWMutex
	h    *mHeap
	lock bool
}

// HeapItem
type HeapItem interface {
	Priority() int64
	SetIndex(int)
	Index() int
}

// NewXHeap create a min or max heap
func NewXHeap() *XHeap {
	h := &mHeap{
		items: []HeapItem{},
	}
	return &XHeap{
		h: h,
	}
}

func (h *XHeap) NoLock() *XHeap {
	h.lock = false
	return h
}

func (h *XHeap) Lock() *XHeap {
	h.lock = true
	return h
}

func (h *XHeap) LockModel() bool {
	return h.lock
}

func (h *XHeap) Push(items ...HeapItem) {
	if h.lock {
		h.l.Lock()
		defer h.l.Unlock()
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		heap.Push(h.h, item)
	}
}

func (h *XHeap) Pop() HeapItem {
	if h.lock {
		h.l.Lock()
		defer h.l.Unlock()
	}
	l := len(h.h.items)
	if l == 0 {
		return nil
	}
	hi := heap.Pop(h.h)
	return hi.(HeapItem)
}

func (h *XHeap) Take(idx ...int) HeapItem {
	i := 0
	if len(idx) > 0 {
		i = idx[0]
	}
	if i < 0 {
		return nil
	}
	if h.lock {
		h.l.RLock()
		defer h.l.RUnlock()
	}
	if i >= len(h.h.items) {
		return nil
	}
	item := h.h.items[i]
	return item
}

func (h *XHeap) Fix(idx int) {
	if idx < 0 {
		return
	}
	if h.lock {
		h.l.Lock()
		defer h.l.Unlock()
	}
	if idx >= len(h.h.items) {
		return
	}
	heap.Fix(h.h, idx)
}

func (h *XHeap) RemoveByIdx(idx int) HeapItem {
	if idx < 0 {
		return nil
	}
	if h.lock {
		h.l.Lock()
		defer h.l.Unlock()
	}
	if idx >= len(h.h.items) {
		return nil
	}
	hi := heap.Remove(h.h, idx)
	return hi.(HeapItem)
}

func (h *XHeap) Len() int {
	if h.lock {
		h.l.RLock()
		defer h.l.RUnlock()
	}
	l := len(h.h.items)
	return l
}

// A mHeap implements heap.Interface and holds Items.
type mHeap struct {
	items []HeapItem
}

func (h mHeap) Len() int {
	return len(h.items)
}

func (h mHeap) Less(i, j int) bool {
	return h.items[i].Priority() < h.items[j].Priority()
}

func (h mHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].SetIndex(i)
	h.items[j].SetIndex(j)
}

func (h *mHeap) Push(x interface{}) {
	n := len(h.items)
	x.(HeapItem).SetIndex(n)
	h.items = append(h.items, x.(HeapItem))
}

func (h *mHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	h.items = old[0 : n-1]
	return item
}
