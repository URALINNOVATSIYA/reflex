package reflex

import (
	"reflect"
	"unsafe"
)

const (
	FlagKindWidth           = 5 // there are 27 kinds
	FlagKindMask    uintptr = 1<<FlagKindWidth - 1
	FlagStickyRO    uintptr = 1 << 5
	FlagEmbedRO     uintptr = 1 << 6
	FlagIndir       uintptr = 1 << 7
	FlagAddr        uintptr = 1 << 8
	FlagMethod      uintptr = 1 << 9
	FlagMethodShift         = 10
	FlagRO          uintptr = FlagStickyRO | FlagEmbedRO
)

func PtrOf(v reflect.Value) unsafe.Pointer {
	rv := reflect.ValueOf(v)
	f := MakeExported(rv.Field(1))
	return f.Interface().(unsafe.Pointer)
}

func FlagOf(v reflect.Value) uintptr {
	rv := reflect.ValueOf(v)
	f := MakeExported(rv.Field(2))
	return uintptr(f.Uint())
}

type Value struct {
	reflect.Value
	T    *AbiType       // reflect.Value.typ_
	Ptr  unsafe.Pointer // reflect.Value.ptr
	Flag uintptr        // reflect.Value.flag
}

func ValueOf(v any) *Value {
	return ValueFrom(reflect.ValueOf(v))
}

func ValueFrom(v reflect.Value) *Value {
	value := &Value{Value: v}
	p := ProxyOf(v)
	value.T = abiTypeFromValue(v)
	value.Ptr = p.Field(1).Interface().(unsafe.Pointer)
	value.Flag = uintptr(p.Field(2).Uint())
	return value
}

func (v *Value) IsKindOf(expected reflect.Kind) bool {
	return reflect.Kind(v.Flag&FlagKindMask) == expected
}

func (v *Value) IsExported() bool {
	return v.Flag&FlagRO == 0
}

func (v *Value) IsAddressable() bool {
	return v.Flag&FlagAddr != 0
}

func (v *Value) IsAssignable() bool {
	return v.IsExported() && v.IsAddressable()
}
