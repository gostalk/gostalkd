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
package structure_test

import (
	"testing"

	"github.com/sjatsh/beanstalkd-go/structure"
)

type TestHeapValue struct {
	A   string
	pri int
}

func TestHeap_Push(t *testing.T) {
	h := structure.NewHeap()
	h.SetLessFn(func(i, j *structure.Item) bool {
		return i.Value.(*TestHeapValue).pri < j.Value.(*TestHeapValue).pri
	})

	h.Push(&structure.Item{
		Value: &TestHeapValue{
			A:   "v1",
			pri: 1,
		},
	})

	h.Push(&structure.Item{
		Value: &TestHeapValue{
			A:   "v2",
			pri: 2,
		},
	})
}

func TestHeap_Pop(t *testing.T) {
	h := structure.NewHeap()
	h.SetLessFn(func(i, j *structure.Item) bool {
		return i.Value.(*TestHeapValue).pri < j.Value.(*TestHeapValue).pri
	})

	h.Push(&structure.Item{
		Value: &TestHeapValue{
			A:   "v1",
			pri: 1,
		},
	})

	h.Push(&structure.Item{
		Value: &TestHeapValue{
			A:   "v2",
			pri: 2,
		},
	})

	v1 := h.Pop()
	v2 := h.Pop()
	t.Logf("value: %#v, index=%d", v1.Value, v1.Index())
	t.Logf("value: %#v, index=%d", v2.Value, v2.Index())
}

func TestHeap_Take(t *testing.T) {
	h := structure.NewHeap()
	h.SetLessFn(func(i, j *structure.Item) bool {
		return i.Value.(*TestHeapValue).pri < j.Value.(*TestHeapValue).pri
	})

	h.Push(&structure.Item{
		Value: &TestHeapValue{
			A:   "v1",
			pri: 1,
		},
	})

	h.Push(&structure.Item{
		Value: &TestHeapValue{
			A:   "v2",
			pri: 2,
		},
	})

	t.Logf("value: %#v, index=%d", h.Take().Value, h.Take().Index())

	v1 := h.Pop()
	t.Logf("value: %#v, index=%d", v1.Value, v1.Index())
	t.Logf("value: %#v, index=%d", h.Take().Value, h.Take().Index())
}

func TestHeap_Set(t *testing.T) {
	h := structure.NewHeap()
	h.SetLessFn(func(i, j *structure.Item) bool {
		return i.Value.(*TestHeapValue).pri < j.Value.(*TestHeapValue).pri
	})
	h.SetPosFn(func(item *structure.Item, i int) {
		t.Logf("value: %#v, index=%d", item, i)
	})

	v1 := &structure.Item{
		Value: &TestHeapValue{
			A:   "v1",
			pri: 1,
		},
	}
	v2 := &structure.Item{
		Value: &TestHeapValue{
			A:   "v2",
			pri: 2,
		},
	}

	h.Push(v1)
	h.Push(v2)

	v1.Value.(*TestHeapValue).A = "v3"

	h.Set(0, v1)

	t.Logf("value: %#v, index=%d", h.Take().Value, h.Take().Index())
}

func TestHeap_Remove(t *testing.T) {
	h := structure.NewHeap()
	h.SetLessFn(func(i, j *structure.Item) bool {
		return i.Value.(*TestHeapValue).pri < j.Value.(*TestHeapValue).pri
	})

	v1 := &structure.Item{
		Value: &TestHeapValue{
			A:   "v1",
			pri: 1,
		},
	}
	v2 := &structure.Item{
		Value: &TestHeapValue{
			A:   "v2",
			pri: 2,
		},
	}
	h.Push(v1)
	h.Push(v2)

	t.Logf("value: %#v", h.Take(0).Value)
	t.Logf("value: %#v", h.Take(1).Value)

	h.Remove(1)

	t.Logf("value: %#v", h.Take(1))
}
