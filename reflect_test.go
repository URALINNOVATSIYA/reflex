package reflex

import (
	"reflect"
	"testing"
	"unsafe"
)

type testRecPtr *testRecPtr
type testInterface any

func TestNameOf(t *testing.T) {
	items := []struct {
		t    reflect.Type
		name string
	}{
		{reflect.TypeOf(nil), "nil"},
		{reflect.TypeOf(false), "bool"},
		{reflect.TypeFor[string](), "string"},
		{reflect.TypeFor[uint8](), "uint8"},
		{reflect.TypeFor[int8](), "int8"},
		{reflect.TypeFor[uint16](), "uint16"},
		{reflect.TypeFor[int16](), "int16"},
		{reflect.TypeFor[uint32](), "uint32"},
		{reflect.TypeFor[int32](), "int32"},
		{reflect.TypeFor[uint64](), "uint64"},
		{reflect.TypeFor[int64](), "int64"},
		{reflect.TypeFor[uint](), "uint"},
		{reflect.TypeFor[int](), "int"},
		{reflect.TypeFor[float32](), "float32"},
		{reflect.TypeFor[float64](), "float64"},
		{reflect.TypeFor[complex64](), "complex64"},
		{reflect.TypeFor[complex128](), "complex128"},
		{reflect.TypeFor[uintptr](), "uintptr"},
		{reflect.TypeFor[unsafe.Pointer](), "unsafe.Pointer"},
		{reflect.TypeFor[[][]int](), "[][]int"},
		{reflect.TypeFor[map[byte][]int](), "map[uint8][]int"},
		{reflect.TypeFor[any](), "interface {}"},
		{reflect.TypeFor[*any](), "*interface {}"},
		{reflect.TypeFor[*bool](), "*bool"},
		{reflect.TypeFor[***complex64](), "***complex64"},
		{reflect.TypeFor[testRecPtr](), "github.com/URALINNOVATSIYA/reflex.testRecPtr"},
		{reflect.TypeFor[testInterface](), "github.com/URALINNOVATSIYA/reflex.testInterface"},
	}
	for i, item := range items {
		actual := NameOf(item.t)
		if item.name != actual {
			t.Errorf("name of type #%d must be %q, but received %q", i+1, item.name, actual)
		}
	}
}
