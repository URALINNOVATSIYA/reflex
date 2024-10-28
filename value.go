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

func DirPtrOf(v reflect.Value) unsafe.Pointer {
	rv := reflect.ValueOf(v)
	ptr := MakeExported(rv.Field(1)).Interface().(unsafe.Pointer)
	flag := uintptr(MakeExported(rv.Field(2)).Uint())
	if flag&FlagIndir != 0 {
		return *(*unsafe.Pointer)(ptr)
	}
	return ptr	
}

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

func ValueOf(v any) Value {
	return ValueFrom(reflect.ValueOf(v))
}

func ValueFrom(v reflect.Value) Value {
	value := Value{Value: v}
	p := ProxyOf(v)
	value.T = abiTypeFromValue(v)
	value.Ptr = p.Field(1).Interface().(unsafe.Pointer)
	value.Flag = uintptr(p.Field(2).Uint())
	return value
}

func (v Value) DirPtr() unsafe.Pointer {
	if v.IsIndirect() {
		return *(*unsafe.Pointer)(v.Ptr)
	}
	return v.Ptr
}

func (v Value) IsKindOf(expected reflect.Kind) bool {
	return reflect.Kind(v.Flag&FlagKindMask) == expected
}

func (v Value) IsExported() bool {
	return v.Flag&FlagRO == 0
}

func (v *Value) IsAddressable() bool {
	return v.Flag&FlagAddr != 0
}

func (v Value) IsAssignable() bool {
	return v.IsExported() && v.IsAddressable()
}

func (v Value) IsIndirect() bool {
	return v.Flag&FlagIndir != 0
}

func (v Value) Elem() Value {
	return ValueFrom(v.Value.Elem())
}

func (v Value) Field(i int) Value {
	return ValueFrom(v.Value.Field(i))
}

func (v Value) FieldByName(name string) Value {
	return ValueFrom(v.Value.FieldByName(name))
}

func (v Value) FieldByNameFunc(match func(string) bool) Value {
	return ValueFrom(v.Value.FieldByNameFunc(match))
}

func (v Value) FieldByIndex(index []int) Value {
	return ValueFrom(v.Value.FieldByIndex(index))
}

func (v Value) FieldByIndexErr(index []int) (Value, error) {
	field, err := v.Value.FieldByIndexErr(index)
	return ValueFrom(field), err
}

func (v Value) Index(i int) Value {
	return ValueFrom(v.Value.Index(i))
}

func (v Value) MapIndex(key reflect.Value) Value {
	return ValueFrom(v.Value.MapIndex(key))
}

func (v Value) MapKeys() []Value {
	keys := v.Value.MapKeys()
	pkeys := make([]Value, len(keys))
	for i, size := 0, len(keys); i < size; i++ {
		pkeys[i] = ValueFrom(keys[i])
	}
	return pkeys
}

func (v Value) MapValues() []Value {
	values := make([]Value, v.Value.Len())
	for i, iter := 0, v.Value.MapRange(); iter.Next(); i++ {
		values[i] = ValueFrom(iter.Value())
	}
	return values
}

func (v Value) Method(i int) Value {
	return ValueFrom(v.Value.Method(i))
}

func (v Value) MethodByName(name string) Value {
	return ValueFrom(v.Value.MethodByName(name))
}

func (v Value) Recv() (x Value, ok bool) {
	value, ok := v.Value.Recv()
	return ValueFrom(value), ok
}