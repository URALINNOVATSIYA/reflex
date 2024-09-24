package reflex

import (
	"reflect"
	"unsafe"
)

type AbiType struct {
	Size       uintptr
	PtrBytes   uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash       uint32  // hash of type; avoids computation in hash tables
	TFlag      uint8   // extra type information flags
	Align      uint8   // alignment of variable with this type
	FieldAlign uint8   // alignment of struct field with this type
	Kind       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// GCData stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, GCData is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	GCData    *byte
	Str       int32 // string form
	PtrToThis int32 // type for pointer to this type, may be zero
}

// abiTypeFromValue converts reflect.Value to reflex.AbiType
func abiTypeFromValue(v reflect.Value) *AbiType {
	t := ProxyOf(v).Field(0).Elem()
	return &AbiType{
		Size:       t.Field(0).Interface().(uintptr),
		PtrBytes:   t.Field(1).Interface().(uintptr),
		Hash:       t.Field(2).Interface().(uint32),
		TFlag:      uint8(t.Field(3).Uint()),
		Align:      t.Field(4).Interface().(uint8),
		FieldAlign: t.Field(5).Interface().(uint8),
		Kind:       t.Field(6).Interface().(uint8),
		Equal:      t.Field(7).Interface().(func(unsafe.Pointer, unsafe.Pointer) bool),
		GCData:     t.Field(8).Interface().(*byte),
		Str:        int32(t.Field(9).Int()),
		PtrToThis:  int32(t.Field(10).Int()),
	}
}
