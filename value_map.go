package reflex

import (
	"reflect"
	"unsafe"
)

type ValueMap struct {
	values map[unsafe.Pointer]int
}

func NewValueMap() *ValueMap {
	return &ValueMap{
		make(map[unsafe.Pointer]int),
	}
}

func (m *ValueMap) Get(v reflect.Value) (int, bool) {
	id, exists := m.values[PtrOf(v)]
	return id, exists
}

func (m *ValueMap) Add(id int, v reflect.Value) bool {
	ptr := PtrOf(v)
	if _, exists := m.values[ptr]; exists {
		return true
	}
	m.values[ptr] = id
	return false
}

func (m *ValueMap) Put(id int, v reflect.Value) {
	m.values[PtrOf(v)] = id
}
