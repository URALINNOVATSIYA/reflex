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

func (s *Slice) IsVirtual() bool {
	return s.Id < 0
}

func (s *Slice) Addr() Addr {
	return Address(s.V)
}

func (s *Slice) Key() SliceKey {
	return SliceKey{
		ElemType:  s.ElemType,
		Ptr:       s.Ptr,
		PtrLenEnd: s.PtrLenEnd,
		PtrCapEnd: s.PtrCapEnd,
	}
}

func (s *Slice) Len() int {
	return int(s.PtrLenEnd-s.Ptr) / int(s.ElemType.Size())
}

func (s *Slice) Cap() int {
	return int(s.PtrCapEnd-s.Ptr) / int(s.ElemType.Size())
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
		return 0, s.Len(), s.Cap()
	}
	if s.Ptr > other.Ptr || s.PtrCapEnd < other.PtrCapEnd || s.PtrLenEnd < other.PtrLenEnd {
		return -1, -1, -1
	}
	elemSize := s.ElemType.Size()
	i := int((other.Ptr - s.Ptr) / elemSize)
	j := int((other.PtrLenEnd - s.Ptr) / elemSize)
	k := max(other.Cap(), j)
	return i, j, k
}

func (s *Slice) addChild(slice *Slice) {
	if s.ElemType != slice.ElemType {
		panic("mismatched element types")
	}
	s.Childs = append(s.Childs, slice)
	slice.Parent = s
}

type SliceMap struct {
	items   map[Addr]*Slice
	idmap   map[int]*Slice
	parents []*Slice
	initId  int
	id      int
}

func NewSliceMap(initialVirtualId int) *SliceMap {
	return &SliceMap{
		items:  make(map[Addr]*Slice),
		idmap:  make(map[int]*Slice),
		initId: initialVirtualId,
		id:     initialVirtualId,
	}
}

func (m *SliceMap) Clear() {
	m.id = m.initId
	m.parents = nil
	clear(m.items)
	clear(m.idmap)
}

func (m *SliceMap) Parents() []*Slice {
	return m.parents
}

func (m *SliceMap) Get(id int) *Slice {
	return m.idmap[id]
}

func (m *SliceMap) GetByValue(v reflect.Value) *Slice {
	return m.items[Address(v)]
}

func (m *SliceMap) Has(v reflect.Value) bool {
	return m.GetByValue(v) != nil
}

func (m *SliceMap) Add(v reflect.Value, id int) (*Slice, bool) {
	if id < 0 {
		panic("idntity must not be negative")
	}
	slice := NewSlice(v, id)
	if m.idmap[slice.Id] != nil {
		panic("slice id must be unique")
	}
	if s := m.items[slice.Addr()]; s != nil {
		if !s.IsVirtual() {
			return s, false
		}
	}
	return slice, m.add(slice)
}

func (m *SliceMap) add(slice *Slice) bool {
	for i, parent := range m.parents {
		switch parent.Relation(slice) {
		case SliceRelationSelf:
			if parent.IsVirtual() {
				delete(m.idmap, parent.Id)
				delete(m.items, parent.Addr())
				parent.Id = slice.Id
				parent.V = slice.V
				m.registerSlice(parent)
				return true
			}
			parent.addChild(slice)
			m.registerSlice(slice)
			return false
		case SliceRelationParent:
			parent.addChild(slice)
			m.registerSlice(slice)
			return true
		case SliceRelationChild:
			m.parents[i] = slice
			for _, child := range parent.Childs {
				slice.addChild(child)
			}
			parent.Childs = nil
			slice.addChild(parent)
			m.registerSlice(slice)
			return true
		case SliceRelationRelative:
			m.registerSlice(slice)
			p := commonParent(parent, slice, m.id)
			m.id--
			for _, child := range parent.Childs {
				p.addChild(child)
			}
			if parent.IsVirtual() {
				delete(m.idmap, parent.Id)
				delete(m.items, parent.Addr())
			} else {
				parent.Childs = nil
				p.addChild(parent)
			}
			p.addChild(slice)
			last := len(m.parents) - 1
			m.parents[i] = m.parents[last]
			m.parents[last] = nil
			m.parents = m.parents[:last]
			return m.add(p)
		}
	}
	m.parents = append(m.parents, slice)
	m.registerSlice(slice)
	return true
}

func (m *SliceMap) registerSlice(slice *Slice) {
	m.items[slice.Addr()] = slice
	m.idmap[slice.Id] = slice
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
