package reflex

import (
	"iter"
	"reflect"
	"unsafe"
)

type chanHeader struct {
	qcount   uintptr
	dataqsiz uintptr
	buf      unsafe.Pointer
	elemsize uint16
	closed   uint32
	timer    unsafe.Pointer
	elemtype unsafe.Pointer
	sendx    uintptr
	recvx    uintptr
}

// ChanElems returns a copy of the buffered channel values in receive order.
// The channel must not be used concurrently while its runtime buffer is read.
func ChanElems(ch reflect.Value) []reflect.Value {
	if ch.Kind() != reflect.Chan {
		panic("argument is not a channel")
	}
	capacity := ch.Cap()
	length := ch.Len()
	if capacity == 0 || length == 0 {
		return nil
	}
	elemType := ch.Type().Elem()
	elemSize := elemType.Size()
	items := make([]reflect.Value, length)
	if elemSize == 0 {
		for i := range items {
			items[i] = reflect.Zero(elemType)
		}
		return items
	}
	h := (*chanHeader)(DataPtrOf(ch))
	idx := int(h.recvx)
	for i := range items {
		ptr := unsafe.Add(h.buf, uintptr(idx)*elemSize)
		items[i] = reflect.NewAt(elemType, ptr).Elem()
		idx++
		if idx == capacity {
			idx = 0
		}
	}
	return items
}

// ChanElemIter iterates over the buffered channel values in receive order.
// The channel must not be used concurrently while its runtime buffer is read.
func ChanElemIter(ch reflect.Value) iter.Seq[reflect.Value] {
	if ch.Kind() != reflect.Chan {
		panic("argument is not a channel")
	}
	return func(yield func(reflect.Value) bool) {
		capacity := ch.Cap()
		length := ch.Len()
		if capacity == 0 || length == 0 {
			return
		}
		elemType := ch.Type().Elem()
		elemSize := elemType.Size()
		if elemSize == 0 {
			for range length {
				if !yield(reflect.Zero(elemType)) {
					return
				}
			}
			return
		}
		h := (*chanHeader)(DataPtrOf(ch))
		idx := int(h.recvx)
		for range length {
			ptr := unsafe.Add(h.buf, uintptr(idx)*elemSize)
			if !yield(reflect.NewAt(elemType, ptr).Elem()) {
				return
			}
			idx++
			if idx == capacity {
				idx = 0
			}
		}
	}
}
