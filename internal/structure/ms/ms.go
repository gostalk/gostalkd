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

package ms

import (
	"sync"
)

type Ms struct {
	last int
	lock bool
	sync.RWMutex

	items []interface{}

	onInsert *EventFn
	onRemove *EventFn
}

type EventFn func(ms *Ms, item interface{}, i int)

// NewMs 创建一个队列
func NewMs(items ...interface{}) *Ms {
	return &Ms{
		items: items,
		lock:  true,
	}
}

func (m *Ms) NoLock() {
	m.lock = false
}

func (m *Ms) WithInsertFn(fn EventFn) *Ms {
	m.onInsert = &fn
	return m
}

func (m *Ms) WithDelFn(fn EventFn) *Ms {
	m.onRemove = &fn
	return m
}

func (m *Ms) Len() int {
	if m.lock {
		m.RLock()
		defer m.RUnlock()
	}
	l := len(m.items)
	return l
}

func (m *Ms) Iterator(fn func(item interface{}) bool) {
	if m.lock {
		m.RLock()
		defer m.RUnlock()
	}
	for _, v := range m.items {
		if ok := fn(v); !ok {
			break
		}
	}
}

func (m *Ms) Append(items ...interface{}) int {
	if m.lock {
		m.Lock()
		defer m.Unlock()
	}
	m.items = append(m.items, items...)
	l := len(m.items)
	if m.onInsert != nil {
		for _, item := range items {
			(*m.onInsert)(m, item, l-1)
		}
	}
	return l
}

func (m *Ms) Take() interface{} {
	if len(m.items) == 0 {
		return nil
	}
	if m.lock {
		m.Lock()
		defer m.Unlock()
	}
	m.last = m.last % len(m.items)
	item := m.items[m.last]
	m.items = append(m.items[:m.last], m.items[m.last+1:]...)
	if m.onRemove != nil {
		(*m.onRemove)(m, item, m.last)
	}
	m.last++
	return item
}

func (m *Ms) DeleteWithIdx(idx int) interface{} {
	if idx < 0 {
		return nil
	}
	if m.lock {
		m.Lock()
		defer m.Unlock()
	}
	if idx >= len(m.items) {
		return nil
	}
	item := m.items[idx]
	m.items = append(m.items[:idx], m.items[idx+1:]...)
	if m.onRemove != nil {
		(*m.onRemove)(m, item, idx)
	}
	return item
}

func (m *Ms) DeleteWithItem(item interface{}) bool {
	if item == nil {
		return false
	}
	if m.lock {
		m.Lock()
		defer m.Unlock()
	}
	for i := 0; i < len(m.items); i++ {
		if item == m.items[i] {
			item := m.items[i]
			m.items = append(m.items[:i], m.items[i+1:]...)
			if m.onRemove != nil {
				(*m.onRemove)(m, item, i)
			}
			return true
		}
	}
	return false
}

func (m *Ms) Contains(item interface{}) bool {
	if m.lock {
		m.RLock()
		defer m.RUnlock()
	}
	for i := 0; i < len(m.items); i++ {
		if item == m.items[i] {
			return true
		}
	}
	return false
}

func (m *Ms) Clean() {
	if m.lock {
		m.Lock()
		defer m.Unlock()
	}
	m.last = 0
	m.items = []interface{}{}
}
