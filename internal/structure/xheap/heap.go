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

// An Item is something we manage in a priority queue.
type Item struct {
	value interface{} // The value of the item; arbitrary.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// NewItem
func NewItem(v interface{}) *Item {
	return &Item{
		value: v,
	}
}

// Index
func (item *Item) Index() int {
	return item.index
}

// Value
func (item *Item) Value() interface{} {
	return item.value
}

// Set
func (item *Item) Set(v interface{}) {
	item.value = v
}

// XHeap
type XHeap struct {
	sync.RWMutex
	h    *mHeap
	lock bool
}

// LessFn
type LessFn func(i, j *Item) bool

// NewXHeap create a min or max heap
func NewXHeap() *XHeap {
	h := &mHeap{
		items: []*Item{},
	}
	return &XHeap{
		h:    h,
		lock: true,
	}
}

func (h *XHeap) NoLock() {
	h.lock = false
}

// SetLessFn
func (h *XHeap) SetLessFn(less LessFn) *XHeap {
	h.h.less = &less
	return h
}

func (h *XHeap) Push(items ...*Item) {
	if h.lock {
		h.Lock()
		defer h.Unlock()
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		heap.Push(h.h, item)
	}
}

func (h *XHeap) Pop() *Item {
	if h.lock {
		h.Lock()
		defer h.Unlock()
	}
	l := len(h.h.items)
	if l == 0 {
		return nil
	}
	hi := heap.Pop(h.h)
	return hi.(*Item)
}

func (h *XHeap) Take(idx ...int) *Item {
	i := 0
	if len(idx) > 0 {
		i = idx[0]
	}
	if i < 0 {
		return nil
	}
	if h.lock {
		h.RLock()
		defer h.RUnlock()
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
		h.Lock()
		defer h.Unlock()
	}
	if idx >= len(h.h.items) {
		return
	}
	heap.Fix(h.h, idx)
}

func (h *XHeap) RemoveWithIdx(idx int) *Item {
	if idx < 0 {
		return nil
	}
	if h.lock {
		h.Lock()
		defer h.Unlock()
	}
	if idx >= len(h.h.items) {
		return nil
	}
	hi := heap.Remove(h.h, idx)
	return hi.(*Item)
}

func (h *XHeap) RemoveWithItem(item *Item) *Item {
	if item == nil || item.index < 0 {
		return nil
	}
	if h.lock {
		h.Lock()
		defer h.Unlock()
	}
	if item.index >= len(h.h.items) {
		return nil
	}
	hi := heap.Remove(h.h, item.index)
	return hi.(*Item)
}

func (h *XHeap) Len() int {
	if h.lock {
		h.RLock()
		defer h.RUnlock()
	}
	l := len(h.h.items)
	return l
}

// A mHeap implements heap.Interface and holds Items.
type mHeap struct {
	items []*Item
	less  *LessFn
}

func (h mHeap) Len() int {
	return len(h.items)
}

func (h mHeap) Less(i, j int) bool {
	return (*h.less)(h.items[i], h.items[j])
}

func (h mHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].index = i
	h.items[j].index = j
}

func (h *mHeap) Push(x interface{}) {
	n := len(h.items)
	x.(*Item).index = n
	h.items = append(h.items, x.(*Item))
}

func (h *mHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	h.items = old[0 : n-1]
	return item
}
