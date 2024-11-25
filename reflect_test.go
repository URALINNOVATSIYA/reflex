package reflex

import (
	"reflect"
	"testing"
	"unsafe"
)

type testRecPtr *testRecPtr
type testInterface interface{}

func TestNameOf(t *testing.T) {
	items := []struct {
		t    reflect.Type
		name string
	}{
		{reflect.TypeOf(nil), "nil"},
		{reflect.TypeOf(false), "bool"},
		{reflect.TypeOf(""), "string"},
		{reflect.TypeOf(uint8(0)), "uint8"},
		{reflect.TypeOf(int8(0)), "int8"},
		{reflect.TypeOf(uint16(0)), "uint16"},
		{reflect.TypeOf(int16(0)), "int16"},
		{reflect.TypeOf(uint32(0)), "uint32"},
		{reflect.TypeOf(int32(0)), "int32"},
		{reflect.TypeOf(uint64(0)), "uint64"},
		{reflect.TypeOf(int64(0)), "int64"},
		{reflect.TypeOf(uint(0)), "uint"},
		{reflect.TypeOf(int(0)), "int"},
		{reflect.TypeOf(float32(0)), "float32"},
		{reflect.TypeOf(float64(0)), "float64"},
		{reflect.TypeOf(complex64(0)), "complex64"},
		{reflect.TypeOf(complex128(0)), "complex128"},
		{reflect.TypeOf(uintptr(0)), "uintptr"},
		{reflect.TypeOf(unsafe.Pointer(nil)), "unsafe.Pointer"},
		{reflect.TypeOf([][]int{}), "[][]int"},
		{reflect.TypeOf(map[byte][]int{}), "map[uint8][]int"},
		{reflect.TypeOf([]any{}).Elem(), "interface {}"},
		{reflect.TypeOf((*any)(nil)), "*interface {}"},
		{reflect.TypeOf((*bool)(nil)), "*bool"},
		{reflect.TypeOf((***complex64)(nil)), "***complex64"},
		{reflect.TypeOf(testRecPtr(nil)), "github.com/URALINNOVATSIYA/reflex.testRecPtr"},
		{reflect.TypeOf([]testInterface{}).Elem(), "github.com/URALINNOVATSIYA/reflex.testInterface"},
	}
	for i, item := range items {
		actual := NameOf(item.t)
		if item.name != actual {
			t.Errorf("name of type #%d must be %q, but received %q", i+1, item.name, actual)
		}
	}
}
