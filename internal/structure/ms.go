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
package structure

type Ms struct {
	last  int
	items []interface{}

	onInsert *MsEventFn
	onRemove *MsEventFn
}

type MsEventFn func(ms *Ms, item interface{}, i int)

// NewMs
func NewMs() *Ms {
	return &Ms{
		items: []interface{}{},
	}
}

func (m *Ms) WithInsertEventFn(fn MsEventFn) *Ms {
	m.onInsert = &fn
	return m
}

func (m *Ms) WithDelEventFn(fn MsEventFn) *Ms {
	m.onRemove = &fn
	return m
}

func (m *Ms) Len() int {
	return len(m.items)
}

func (m *Ms) Iterator(fn func(item interface{}) (bool, error)) {
	for _, v := range m.items {
		if bk, err := fn(v); err != nil || bk {
			break
		}
	}
}

func (m *Ms) Append(items ...interface{}) int {
	m.items = append(m.items, items...)
	l := len(m.items)
	if m.onInsert != nil {
		for _, item := range items {
			(*m.onInsert)(m, item, l-1)
		}
	}
	return l
}

func (m *Ms) Delete(idx int) interface{} {
	if idx < 0 || idx >= len(m.items) {
		return nil
	}
	item := m.items[idx]
	m.items = append(m.items[:idx], m.items[idx+1:])
	if m.onRemove != nil {
		(*m.onRemove)(m, item, idx)
	}
	return item
}

func (m *Ms) Clear() {
	m.last = 0
	m.items = []interface{}{}
}

func (m *Ms) Remove(item interface{}) bool {
	for i := 0; i < len(m.items); i++ {
		if item == m.items[i] {
			m.Delete(i)
			return true
		}
	}
	return false
}

func (m *Ms) Contains(item interface{}) bool {
	for i := 0; i < len(m.items); i++ {
		if item == m.items[i] {
			return true
		}
	}
	return false
}

func (m *Ms) Take() interface{} {
	if len(m.items) == 0 {
		return nil
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
