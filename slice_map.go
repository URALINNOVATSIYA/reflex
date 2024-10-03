package reflex

import (
	"fmt"
	"reflect"
)

type SliceRelation byte

const (
	SliceRelationNone SliceRelation = iota
	SliceRelationSelf
	SliceRelationParent
	SliceRelationChild
	SliceRelationRelative
)

type SliceMap struct {
	slices []*sliceDetails
}

func NewSliceMap() *SliceMap {
	return &SliceMap{}
}

type sliceDetails struct {
	slice        reflect.Value
	len          int
	cap          int
	itemSize     uintptr
	firstItemPtr uintptr
	t            reflect.Kind
}

type BoundSlice struct {
	Slice        reflect.Value
	intersectLen int
	Low          int
	High         int
	Max          int
	Relation     SliceRelation
	Type         reflect.Kind
}

func (sm *SliceMap) Add(slice reflect.Value) {
	if slice.Type().Kind() != reflect.Slice && (slice.Type().Kind() == reflect.Pointer && slice.Elem().Type().Kind() != reflect.Array) {
		panic(fmt.Sprintf("unexpected type of slice: %s", slice.Type().Kind().String()))
	}

	sm.slices = append(sm.slices, sm.sliceDetails(slice))
}

func (sm *SliceMap) sliceDetails(slice reflect.Value) *sliceDetails {
	t := slice.Type().Kind()
	if t == reflect.Pointer {
		t = slice.Elem().Type().Kind()
	}

	return &sliceDetails{
		slice:        slice,
		len:          slice.Len(),
		cap:          slice.Cap(),
		itemSize:     sm.sliceItemSize(slice),
		firstItemPtr: slice.Pointer(),
		t:            t,
	}
}

func (sm *SliceMap) sliceItemSize(slice reflect.Value) uintptr {
	if slice.Type().Kind() == reflect.Slice {
		if slice.Len() > 0 {
			return slice.Index(0).Type().Size()
		}

		return 0
	}

	if slice.Elem().Len() > 0 {
		slice.Elem().Index(0).Type().Size()
	}

	return 0
}

func (sm *SliceMap) Find(slice reflect.Value) []BoundSlice {
	if slice.Type().Kind() != reflect.Slice {
		panic(fmt.Sprintf("unexpected type of slice: %s", slice.Type().Kind().String()))
	}

	var boundSlices []BoundSlice

	sDetails := sm.sliceDetails(slice)
	for _, s := range sm.slices {
		if (sDetails.len > 0 && s.len > 0) && sDetails.itemSize != s.itemSize {
			continue
		}

		count, l, h, m, r := sm.intersect(sDetails, s)
		if r == SliceRelationNone {
			continue
		}

		boundSlice := BoundSlice{
			intersectLen: count,
			Slice:        s.slice,
			Low:          l,
			High:         h,
			Max:          m,
			Relation:     r,
			Type:         s.t,
		}

		for i, bSlice := range boundSlices {
			if bSlice.intersectLen > boundSlice.intersectLen {
				continue
			}

			boundSlices[i], boundSlice = boundSlice, bSlice
		}

		boundSlices = append(boundSlices, boundSlice)
	}

	return boundSlices
}

func (sm *SliceMap) intersect(slice1 *sliceDetails, slice2 *sliceDetails) (count int, low int, high int, max int, relation SliceRelation) {
	if slice1.firstItemPtr == slice2.firstItemPtr {
		if slice1.len > 0 && slice2.len > 0 {
			if slice1.len > slice2.len {
				return slice1.len - (slice1.len - slice2.len), 0, slice2.len, slice2.cap, SliceRelationChild
			} else if slice2.len > slice1.len {
				return slice2.len - (slice2.len - slice1.len), 0, slice1.len, slice1.cap, SliceRelationParent
			}

			return slice1.len, 0, slice1.len, slice1.cap, SliceRelationSelf
		}

		if slice1.len == 0 && slice2.len == 0 {
			return 0, 0, 0, slice1.cap, SliceRelationSelf
		}

		if slice1.len == 0 {
			return 0, 0, 0, slice1.cap, SliceRelationParent
		}

		return 0, 0, 0, slice1.cap, SliceRelationChild
	}

	if slice1.firstItemPtr < slice2.firstItemPtr {
		if slice1.firstItemPtr+slice1.itemSize*uintptr(slice1.len) <= slice2.firstItemPtr {
			return -1, -1, -1, -1, SliceRelationNone
		}

		count = sm.intersectCount(slice1, slice2)
		return count, 0, count, count, SliceRelationRelative
	}

	if slice2.firstItemPtr+slice2.itemSize*uintptr(slice2.len) <= slice1.firstItemPtr {
		return -1, -1, -1, -1, SliceRelationNone
	}

	count = sm.intersectCount(slice2, slice1)
	low = int((slice1.firstItemPtr - slice2.firstItemPtr) / slice2.itemSize)
	return count, low, low + slice1.len, low + slice1.cap, SliceRelationParent
}

func (sm *SliceMap) intersectCount(slice1 *sliceDetails, slice2 *sliceDetails) int {
	count := int((slice1.firstItemPtr + slice1.itemSize*uintptr(slice1.len) - slice2.firstItemPtr) / slice1.itemSize)

	if count > slice2.len {
		return slice2.len
	}

	return count
}
