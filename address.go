package reflex

import (
	"reflect"
	"unsafe"
)

type Addr struct {
	Ptr  unsafe.Pointer
	Type reflect.Type
}

func (a Addr) IsValid() bool {
	return a.Ptr != nil
}

func Address(v reflect.Value) Addr {
	if !v.IsValid() {
		return Addr{}
	}
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Struct, reflect.Array, reflect.Slice:
		return Addr{
			Ptr:  PtrOf(v),
			Type: v.Type(),
		}
	case reflect.String, reflect.Chan, reflect.Func, reflect.Map:
		return Addr{
			Ptr:  DataPtrOf(v),
			Type: v.Type(),
		}
	}
	return Addr{}
}
