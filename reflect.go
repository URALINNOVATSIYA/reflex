package reflex

import (
	"fmt"
	"reflect"
	"runtime"
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

var flagOffset uintptr

func SetFlag(v reflect.Value, flag uintptr) reflect.Value {
	f := (*uintptr)(unsafe.Add(unsafe.Pointer(&v), flagOffset))
	*f |= uintptr(flag)
	return v
}

func ResetFlag(v reflect.Value, flag uintptr) reflect.Value {
	f := (*uintptr)(unsafe.Add(unsafe.Pointer(&v), flagOffset))
	*f &= ^uintptr(flag)
	return v
}

func MakeExported(v reflect.Value) reflect.Value {
	return ResetFlag(v, FlagRO)
}

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

// Zero returns a reflect.Value representing the assignable zero value for the specified type.
func Zero(t reflect.Type) reflect.Value {
	if t == nil {
		return reflect.Value{}
	}
	return reflect.New(t).Elem()
}

func CopyOf(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return reflect.Value{}
	}
	copy := Zero(v.Type())
	copy.Set(v)
	return copy
}

func PtrTo(t reflect.Type, v reflect.Value) reflect.Value {
	p := reflect.New(t)
	p.Elem().Set(v)
	return p
}

func PtrAt(t reflect.Type, v reflect.Value) reflect.Value {
	return reflect.NewAt(t, unsafe.Pointer(v.UnsafeAddr()))
}

func NameOf(t reflect.Type) string {
	if t == nil {
		return "nil"
	}
	if t.Name() != "" {
		name := t.Name()
		if t.PkgPath() != "" {
			name = t.PkgPath() + "." + name
		}
		return name
	}
	switch t.Kind() {
	case reflect.Invalid:
		return "<nil>"
	case reflect.Pointer:
		return "*" + NameOf(t.Elem())
	case reflect.Slice:
		return "[]" + NameOf(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), NameOf(t.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", NameOf(t.Key()), NameOf(t.Elem()))
	case reflect.Struct:
		f := ""
		for i, numFields := 0, t.NumField(); i < numFields; i++ {
			field := t.Field(i)
			if i > 0 {
				f += "; "
			}
			f += fmt.Sprintf("%s %s", field.Name, NameOf(t.Field(i).Type))
		}
		return fmt.Sprintf("struct { %s }", f)
	case reflect.Chan:
		switch t.ChanDir() {
		case reflect.BothDir:
			return "chan " + NameOf(t.Elem())
		case reflect.RecvDir:
			return "<-chan" + NameOf(t.Elem())
		default:
			return "chan<-" + NameOf(t.Elem())
		}
	}
	return t.String()
}

func FuncNameOf(v reflect.Value) string {
	if v.Kind() != reflect.Func {
		return ""
	}
	if fn := runtime.FuncForPC(v.Pointer()); fn != nil {
		if name := fn.Name(); name != "" {
			return name
		}
	}
	return NameOf(v.Type())
}
