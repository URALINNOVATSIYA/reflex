package reflex

import (
	"reflect"
	"unsafe"
)

var flagOffset uintptr

func MakeExported(v reflect.Value) reflect.Value {
	flag := (*uintptr)(unsafe.Add(unsafe.Pointer(&v), flagOffset))
	*flag &= ^uintptr(FlagRO)
	return v
}

type Proxy struct {
	reflect.Value
}

func ProxyOf(v any) Proxy {
	return Proxy{reflect.ValueOf(v)}
}

func (p Proxy) Elem() reflect.Value {
	return MakeExported(p.Value.Elem())
}

func (p Proxy) Field(i int) reflect.Value {
	return MakeExported(p.Value.Field(i))
}

func (p Proxy) FieldByName(name string) reflect.Value {
	return MakeExported(p.Value.FieldByName(name))
}

func (p Proxy) FieldByNameFunc(match func(string) bool) reflect.Value {
	return MakeExported(p.Value.FieldByNameFunc(match))
}

func (p Proxy) FieldByIndex(index []int) reflect.Value {
	return MakeExported(p.Value.FieldByIndex(index))
}

func (p Proxy) FieldByIndexErr(index []int) (reflect.Value, error) {
	field, err := p.Value.FieldByIndexErr(index)
	return MakeExported(field), err
}

func (p Proxy) Index(i int) reflect.Value {
	return MakeExported(p.Value.Index(i))
}

func (p Proxy) MapIndex(key reflect.Value) reflect.Value {
	return MakeExported(p.Value.MapIndex(key))
}

func (p Proxy) MapKeys() []reflect.Value {
	keys := p.Value.MapKeys()
	for i, size := 0, len(keys); i < size; i++ {
		keys[i] = MakeExported(keys[i])
	}
	return keys
}

func (p Proxy) MapValues() []reflect.Value {
	values := make([]reflect.Value, p.Value.Len())
	for i, iter := 0, p.Value.MapRange(); iter.Next(); i++ {
		values[i] = MakeExported(iter.Value())
	}
	return values
}

func (p Proxy) Method(i int) reflect.Value {
	return MakeExported(p.Value.Method(i))
}

func (p Proxy) MethodByName(name string) reflect.Value {
	return MakeExported(p.Value.MethodByName(name))
}

func (p Proxy) Recv() (x reflect.Value, ok bool) {
	v, ok := p.Value.Recv()
	return MakeExported(v), ok
}
