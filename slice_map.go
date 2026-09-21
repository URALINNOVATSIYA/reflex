package reflex

import (
	"reflect"
	"unsafe"
)

type SliceRelation byte

const (
	SliceRelationNone SliceRelation = iota
	SliceRelationSelf
	SliceRelationParent
	SliceRelationChild
	SliceRelationRelative
)

func (r SliceRelation) String() string {
	switch r {
	case SliceRelationNone:
		return "none"
	case SliceRelationSelf:
		return "self"
	case SliceRelationParent:
		return "parent"
	case SliceRelationChild:
		return "child"
	case SliceRelationRelative:
		return "relative"
	}
	return "unknown"
}

type SliceKey struct {
	ElemType  reflect.Type
	Ptr       uintptr
	PtrLenEnd uintptr
	PtrCapEnd uintptr
}

type Slice struct {
	Id        int
	V         reflect.Value
	ElemType  reflect.Type
	Ptr       uintptr
	PtrLenEnd uintptr
	PtrCapEnd uintptr
	Parent    *Slice
	Childs    []*Slice
}

func NewSlice(v reflect.Value, id int) *Slice {
	var ptr uintptr
	switch v.Kind() {
	case reflect.Slice:
		ptr = uintptr(DirPtrOf(v))
	case reflect.Array:
		ptr = uintptr(PtrOf(v))
	case reflect.Pointer:
		if !v.IsNil() {
			return NewSlice(v.Elem(), id)
		}
		fallthrough
	default:
		panic("invalid argument type")
	}
	elemType := v.Type().Elem()
	elemSize := elemType.Size()
	return &Slice{
		Id:        id,
		V:         v,
		ElemType:  elemType,
		Ptr:       ptr,
		PtrLenEnd: ptr + elemSize*uintptr(v.Len()),
		PtrCapEnd: ptr + elemSize*uintptr(v.Cap()),
	}
}

func (s *Slice) Key() SliceKey {
	return SliceKey{
		ElemType:  s.ElemType,
		Ptr:       s.Ptr,
		PtrLenEnd: s.PtrLenEnd,
		PtrCapEnd: s.PtrCapEnd,
	}
}

func (s *Slice) Relation(other *Slice) SliceRelation {
	if s.ElemType != other.ElemType {
		return SliceRelationNone
	}
	if other.Ptr >= s.PtrCapEnd || other.PtrCapEnd <= s.Ptr {
		return SliceRelationNone
	}
	if s.Ptr == other.Ptr && s.PtrCapEnd == other.PtrCapEnd && s.PtrLenEnd == other.PtrLenEnd {
		return SliceRelationSelf
	}
	if s.Ptr <= other.Ptr && s.PtrCapEnd >= other.PtrCapEnd {
		if s.PtrLenEnd >= other.PtrLenEnd {
			return SliceRelationParent
		}
		if s.Ptr >= other.Ptr {
			return SliceRelationChild
		}
		return SliceRelationRelative
	}
	if other.PtrLenEnd >= s.PtrLenEnd {
		return SliceRelationChild
	}
	return SliceRelationRelative
}

func (s *Slice) SliceOf(other *Slice) (int, int, int) {
	if s.ElemType != other.ElemType {
		return -1, -1, -1
	}
	if other.Ptr >= s.PtrCapEnd || other.PtrCapEnd <= s.Ptr {
		return -1, -1, -1
	}
	if s.Ptr == other.Ptr && s.PtrCapEnd == other.PtrCapEnd && s.PtrLenEnd == other.PtrLenEnd {
		return 0, s.V.Len(), s.V.Cap()
	}
	if s.Ptr > other.Ptr || s.PtrCapEnd < other.PtrCapEnd || s.PtrLenEnd < other.PtrLenEnd {
		return -1, -1, -1
	}
	elemSize := s.ElemType.Size()
	i := (other.Ptr - s.Ptr) / elemSize
	j := (other.PtrLenEnd - s.Ptr) / elemSize
	return int(i), int(j), other.V.Cap()
}

func (s *Slice) addChild(slice *Slice) {
	if s.ElemType != slice.ElemType {
		panic("mismatched element types")
	}
	s.Childs = append(s.Childs, slice)
	slice.Parent = s
}

type SliceMap struct {
	items   map[SliceKey]*Slice
	idmap   map[int]*Slice
	parents []*Slice
	id      int
}

func NewSliceMap() *SliceMap {
	return &SliceMap{
		items: make(map[SliceKey]*Slice),
		idmap: make(map[int]*Slice),
	}
}

func (m *SliceMap) Clear() {
	m.id = 0
	m.parents = nil
	clear(m.items)
	clear(m.idmap)
}

func (m *SliceMap) Get(id int) *Slice {
	return m.idmap[id]
}

func (m *SliceMap) Has(v reflect.Value) bool {
	return m.items[NewSlice(v, 0).Key()] != nil
}

func (m *SliceMap) Add(v reflect.Value, id int) bool {
	if id < 0 {
		panic("idntity must not be negative")
	}
	slice := NewSlice(v, id)
	return m.add(slice)
}

func (m *SliceMap) add(slice *Slice) bool {
	if m.idmap[slice.Id] != nil {
		panic("slice id must be unique")
	}
	key := slice.Key()
	if s := m.items[key]; s != nil {
		if s.Id >= 0 {
			return false
		}
		delete(m.idmap, s.Id)
		s.Id = slice.Id
		s.V = slice.V
		m.idmap[s.Id] = s
		return true
	}
	m.items[key] = slice
	m.idmap[slice.Id] = slice
	for i, parent := range m.parents {
		switch parent.Relation(slice) {
		case SliceRelationParent:
			parent.addChild(slice)
			return true
		case SliceRelationChild:
			m.parents[i] = slice
			for _, child := range parent.Childs {
				slice.addChild(child)
			}
			parent.Childs = nil
			slice.addChild(parent)
			return true
		case SliceRelationRelative:
			m.id--
			p := commonParent(parent, slice, m.id)
			for _, child := range parent.Childs {
				p.addChild(child)
			}
			parent.Childs = nil
			p.addChild(parent)
			p.addChild(slice)
			last := len(m.parents) - 1
			m.parents[i] = m.parents[last]
			m.parents[last] = nil
			m.parents = m.parents[:last]
			return m.add(p)
		}
	}
	m.parents = append(m.parents, slice)
	return true
}

func commonParent(slice1, slice2 *Slice, id int) *Slice {
	if slice1.Ptr > slice2.Ptr {
		slice1, slice2 = slice2, slice1
	}

	elemSize := slice1.ElemType.Size()
	length := (max(slice1.PtrLenEnd, slice2.PtrLenEnd) - slice1.Ptr) / elemSize
	capacity := (max(slice1.PtrCapEnd, slice2.PtrCapEnd) - slice1.Ptr) / elemSize

	data := unsafe.Slice((*byte)(unsafe.Pointer(slice1.Ptr)), int(capacity*elemSize))
	value := reflect.NewAt(reflect.SliceOf(slice1.ElemType), unsafe.Pointer(&data)).Elem()
	value = value.Slice3(0, int(length), int(capacity))

	// Deprecated implementation
	//value := reflect.New(reflect.SliceOf(slice1.ElemType))
	//header := (*reflect.SliceHeader)(unsafe.Pointer(value.Pointer()))
	//header.Data = slice1.Ptr
	//header.Len = int(length)
	//header.Cap = int(capacity)

	return NewSlice(value, id)
}
