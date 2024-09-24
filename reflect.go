package reflex

import (
	"fmt"
	"reflect"
	"runtime"
)

// Zero returns a Value representing the assignable zero value for the specified type.
func Zero(t reflect.Type) reflect.Value {
	if t == nil {
		return reflect.Value{}
	}
	return reflect.New(t).Elem()
}

func PtrTo(t reflect.Type, v reflect.Value) reflect.Value {
	p := reflect.New(t)
	p.Elem().Set(v)
	return p
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
