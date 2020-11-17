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
	"fmt"
	"testing"

	"github.com/gostalk/gostalkd/structure"
)

type TestMs struct {
	A string
}

func TestNewMs(t *testing.T) {
	structure.NewMs()
}

func TestMs_WithInsertEventFn(t *testing.T) {
	ms := structure.NewMs().WithInsertEventFn(func(ms *structure.Ms, item interface{}, i int) {
		fmt.Printf("%+v, %+v, %d", *ms, item, i)
	})
	ms.Append(&TestMs{A: "a"})
}

func TestMs_WithDelEventFn(t *testing.T) {
	ms := structure.NewMs().WithDelEventFn(func(ms *structure.Ms, item interface{}, i int) {
		fmt.Printf("%+v, %+v, %d\n", *ms, item, i)
	})
	ms.Append(TestMs{A: "a"})
	ms.Append(TestMs{A: "b"})

	ms.Delete(0)

	t.Logf("%#v\n", ms.Take())
}

func TestMs_Len(t *testing.T) {
	m := structure.NewMs()
	m.Append(TestMs{
		A: "a",
	})
	t.Log(m.Len())
}

func TestMs_Clear(t *testing.T) {
	m := structure.NewMs()
	m.Append(TestMs{
		A: "a",
	})
	t.Log(m.Take())

	m.Clear()
	t.Log(m.Take())
}

func TestMs_Contains(t *testing.T) {
	m := structure.NewMs()
	t1 := TestMs{
		A: "a",
	}
	t2 := &TestMs{
		A: "b",
	}
	m.Append(t1, t2)
	t.Log(m.Contains(t1))
	t.Log(m.Contains(t2))

	fn := func(tm TestMs) {
		t.Log(m.Contains(tm))
	}
	fn(t1)
}

func TestMs_Delete(t *testing.T) {
	m := structure.NewMs()
	t1 := TestMs{
		A: "a",
	}
	m.Append(t1)
	m.Delete(0)

	t.Log(m.Contains(t1))
}

func TestMs_Remove(t *testing.T) {
	m := structure.NewMs()
	t1 := TestMs{
		A: "a",
	}
	m.Append(t1)
	m.Remove(t1)

	t.Log(m.Contains(t1))
}

func TestMs_Iterator(t *testing.T) {
	m := structure.NewMs()
	t1 := TestMs{A: "a"}
	t2 := TestMs{A: "b"}
	m.Append(t1, t2)

	m.Iterator(func(item interface{}) (bool, error) {
		t.Log(item)
		return false, nil
	})
}

func TestMs_Take(t *testing.T) {
	m := structure.NewMs()
	t1 := TestMs{A: "a"}
	t2 := TestMs{A: "b"}
	m.Append(t1, t2)

	t.Log(m.Take())
	t.Log(m.Take())
}
