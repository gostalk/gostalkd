// Copyright 2020 gostalkd
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

package structure

import (
	"container/heap"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value interface{} // The value of the item; arbitrary.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// Index
func (item *Item) Index() int {
	return item.index
}

type Heap struct {
	heap *mHeap
}

type lessFn func(i, j *Item) bool
type setPosFn func(item *Item, i int)

// NewHeap 创建一个小顶堆
func NewHeap() *Heap {
	h := &mHeap{
		items: []*Item{},
	}
	heap.Init(h)
	return &Heap{
		heap: h,
	}
}

// SetLessFn
func (h *Heap) SetLessFn(less lessFn) {
	h.heap.less = &less
}

// SetPosFn
func (h *Heap) SetPosFn(setPos setPosFn) {
	h.heap.setPos = &setPos
}

// Len
func (h *Heap) Len() int {
	return len(h.heap.items)
}

// Set
func (h *Heap) Set(k int, i *Item) {
	if h.heap.setPos == nil {
		return
	}
	if k < 0 || k >= len(h.heap.items) {
		return
	}
	h.heap.items[k].Value = i.Value
	if h.heap.setPos != nil {
		(*h.heap.setPos)(i, k)
	}
}

func (h *Heap) Push(item *Item) bool {
	heap.Push(h.heap, item)
	return true
}

func (h *Heap) Pop() *Item {
	hi := heap.Pop(h.heap)
	if hi == nil {
		return nil
	}
	return hi.(*Item)
}

func (h *Heap) Take(idx ...int) *Item {
	i := 0
	if len(idx) > 0 {
		i = idx[0]
	}
	if i < 0 || i >= len(h.heap.items) {
		return nil
	}
	return h.heap.items[i]
}

func (h *Heap) Remove(idx int) *Item {
	hi := heap.Remove(h.heap, idx)
	if hi == nil {
		return nil
	}
	return hi.(*Item)
}

// A mHeap implements heap.Interface and holds Items.
type mHeap struct {
	items []*Item

	less   *lessFn
	setPos *setPosFn
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
	item, ok := x.(*Item)
	if !ok {
		return
	}
	n := len(h.items)
	item.index = n
	h.items = append(h.items, item)
}

func (h *mHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	h.items = old[0 : n-1]
	return item
}
