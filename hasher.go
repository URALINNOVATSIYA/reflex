package reflex

/*import (
	"crypto/sha256"
	"encoding/binary"
	"reflect"
)

func Hash(v any) uint64 {
	h := NewHasher()
	h.Add(v)
	return h.Hash()
}

// Hasher hashes values by canonicalizing their object graph.
type Hasher struct {
	roots []reflect.Value
}

// NewHasher creates a graph-based hasher.
func NewHasher() *Hasher {
	return &Hasher{}
}

// Add appends a root value to the graph being hashed.
func (h *Hasher) Add(v any) {
	h.roots = append(h.roots, reflect.ValueOf(v))
}

// Reset releases the roots accumulated by Add.
func (h *Hasher) Reset() {
	clear(h.roots)
	h.roots = h.roots[:0]
}

// Hash returns a deterministic hash of the canonical graph representation.
func (h *Hasher) Hash() uint64 {
	builder := NewGraphBuilder(h.roots)
	digest := sha256.Sum256(builder.Encode())
	return binary.LittleEndian.Uint64(digest[:8])
}*/

import (
	"hash/maphash"
	"math"
	"math/bits"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type mapEntry struct {
	key   reflect.Value
	value reflect.Value
	hash  uint64
	path  string
}

type mapEntryList []mapEntry

func (l mapEntryList) Sort() {
	slices.SortFunc(l, func(k1, k2 mapEntry) int {
		if k1.hash < k2.hash {
			return -1
		}
		if k1.hash > k2.hash {
			return 1
		}
		if k1.path < k2.path {
			return -1
		}
		if k1.path > k2.path {
			return 1
		}
		return 0
	})
}

type valueKey struct {
	Addr Addr
	Kind byte // 0 - pointers, 1 - values
}

type sliceKey struct {
	ElemType reflect.Type
	Ptr      uintptr
	Len      int
	Cap      int
}

func (k sliceKey) capacityEnd() uintptr {
	return k.Ptr + uintptr(k.Cap)*k.ElemType.Size()
}

func newSliceKey(v reflect.Value) sliceKey {
	return sliceKey{
		ElemType: v.Type().Elem(),
		Ptr:      uintptr(DataPtrOf(v)),
		Len:      v.Len(),
		Cap:      v.Cap(),
	}
}

type sliceBacking struct {
	group  uintptr
	offset uintptr
}

type sliceOccurrence struct {
	key      sliceKey
	path     string
	typeName string
}

type graphMetadata struct {
	paths       map[valueKey][]string
	pathBuffers [][]string
	occurrences []sliceOccurrence
	active      map[valueKey]bool
}

func (m *graphMetadata) reset() {
	for key, paths := range m.paths {
		clear(paths)
		m.pathBuffers = append(m.pathBuffers, paths[:0])
		delete(m.paths, key)
	}
	clear(m.active)
	clear(m.occurrences)
	m.occurrences = m.occurrences[:0]
}

func (m *graphMetadata) addPath(key valueKey, path []byte) {
	if m.paths == nil {
		m.paths = make(map[valueKey][]string)
	}
	paths, exists := m.paths[key]
	if !exists && len(m.pathBuffers) > 0 {
		last := len(m.pathBuffers) - 1
		paths = m.pathBuffers[last][:0]
		m.pathBuffers[last] = nil
		m.pathBuffers = m.pathBuffers[:last]
	}
	m.paths[key] = append(paths, string(path))
}

func (m *graphMetadata) enter(key valueKey) bool {
	if m.active == nil {
		m.active = make(map[valueKey]bool)
	}
	if m.active[key] {
		return false
	}
	m.active[key] = true
	return true
}

func (m *graphMetadata) collect(v reflect.Value, path []byte) {
	switch v.Kind() {
	case reflect.Interface:
		if !v.IsNil() {
			m.collect(v.Elem(), m.appendMetadataPath(path, "/i", -1))
		}
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		key := valueKey{Addr: Address(v), Kind: 1}
		m.addPath(key, path)
		if m.enter(key) {
			m.collect(v.Elem(), m.appendMetadataPath(path, "/p", -1))
			delete(m.active, key)
		}
	case reflect.Chan, reflect.UnsafePointer:
		if !v.IsNil() {
			m.addPath(valueKey{Addr: Address(v), Kind: 1}, path)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			m.collect(v.Field(i), m.appendMetadataPath(path, "/f", i))
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			m.collect(v.Index(i), m.appendMetadataPath(path, "/a", i))
		}
	case reflect.Slice:
		if v.IsNil() {
			return
		}
		m.occurrences = append(m.occurrences, sliceOccurrence{
			key:      newSliceKey(v),
			path:     string(path),
			typeName: NameOf(v.Type().Elem()),
		})
		key := valueKey{Addr: Address(v), Kind: 1}
		if m.enter(key) {
			for i := 0; i < v.Len(); i++ {
				m.collect(v.Index(i), m.appendMetadataPath(path, "/s", i))
			}
			delete(m.active, key)
		}
	case reflect.Map:
		if v.IsNil() {
			return
		}
		key := valueKey{Addr: Address(v), Kind: 1}
		if m.enter(key) {
			iter := v.MapRange()
			for iter.Next() {
				m.collect(iter.Key(), m.appendMetadataPath(path, "/mk", -1))
				m.collect(iter.Value(), m.appendMetadataPath(path, "/mv", -1))
			}
			delete(m.active, key)
		}
	}
}

func (m *graphMetadata) appendMetadataPath(path []byte, segment string, index int) []byte {
	path = append(path, segment...)
	if index >= 0 {
		path = strconv.AppendInt(path, int64(index), 10)
		path = append(path, ';')
	}
	return path
}

func typeHasMetadata(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.UnsafePointer, reflect.Chan, reflect.Slice, reflect.Map:
		return true
	case reflect.Struct:
		for field := range t.Fields() {
			if typeHasMetadata(field.Type) {
				return true
			}
		}
	case reflect.Array:
		return typeHasMetadata(t.Elem())
	}
	return false
}

func Hash(v any) uint64 {
	h := NewHasher()
	h.Add(v)
	return h.Hash()
}

// Hasher computes a hash of value structure.
type Hasher struct {
	base          *Hasher
	roots         []reflect.Value
	values        map[valueKey]int
	sliceValues   map[sliceKey]int
	sliceBackings map[sliceKey]sliceBacking
	backingIds    map[uintptr]int
	orders        map[valueKey]string
	metadata      graphMetadata
	mhash         *maphash.Hash
	idx           int
	dirty         bool
}

func NewHasher() *Hasher {
	return &Hasher{
		values:        make(map[valueKey]int),
		sliceValues:   make(map[sliceKey]int),
		sliceBackings: make(map[sliceKey]sliceBacking),
		backingIds:    make(map[uintptr]int),
		orders:        make(map[valueKey]string),
		mhash:         &maphash.Hash{},
	}
}

func (h *Hasher) Reset() {
	h.idx = 0
	h.dirty = false
	h.mhash.Reset()
	h.metadata.reset()
	clear(h.roots)
	h.roots = h.roots[:0]
	clear(h.values)
	clear(h.sliceValues)
	clear(h.sliceBackings)
	clear(h.backingIds)
	clear(h.orders)
}

func (h *Hasher) Hash() uint64 {
	if h.dirty {
		h.build()
	}
	return h.mhash.Sum64()
}

func (h *Hasher) Add(v any) {
	h.roots = append(h.roots, reflect.ValueOf(v))
	h.dirty = true
}

func (h *Hasher) build() {
	h.idx = 0
	if h.base != nil {
		h.idx = h.base.idx
	}
	h.mhash.Reset()
	clear(h.values)
	clear(h.sliceValues)
	clear(h.sliceBackings)
	clear(h.backingIds)
	clear(h.orders)
	h.metadata.reset()
	if len(h.roots) == 0 {
		h.dirty = false
		return
	}
	h.setSeed(uint64(h.roots[0].Kind()))
	h.prepareMetadata(h.roots)
	for _, root := range h.roots {
		h.visit(root)
	}
	h.dirty = false
}

func (h *Hasher) setSeed(seed uint64) {
	s := &maphash.Seed{}
	v := reflect.ValueOf(s).Elem().Field(0)
	v = MakeExported(v)
	if seed == 0 {
		seed--
	}
	v.SetUint(seed)
	h.mhash.SetSeed(*s)
}

func (h *Hasher) prepareMetadata(roots []reflect.Value) {
	clear(h.orders)
	clear(h.sliceBackings)
	h.metadata.reset()
	for index, root := range roots {
		if !root.IsValid() || !typeHasMetadata(root.Type()) {
			continue
		}
		path := h.metadata.appendMetadataPath(nil, "/r", index)
		h.metadata.collect(root, path)
	}
	if len(h.metadata.paths) > 0 {
		if h.orders == nil {
			h.orders = make(map[valueKey]string, len(h.metadata.paths))
		}
		for key, paths := range h.metadata.paths {
			slices.Sort(paths)
			var order strings.Builder
			for i, path := range paths {
				if i > 0 {
					order.WriteByte(0)
				}
				order.WriteString(path)
			}
			h.orders[key] = order.String()
		}
	}
	occurrences := h.metadata.occurrences
	if len(occurrences) == 0 {
		return
	}
	if h.sliceBackings == nil {
		h.sliceBackings = make(map[sliceKey]sliceBacking, len(occurrences))
	}
	h.prepareSliceBackings(occurrences)
}

func (h *Hasher) visit(v reflect.Value) {
	h.idx++
	h.hashIndex(h.idx)
	h.hashType(v)
	switch v.Kind() {
	case reflect.Invalid:
		h.hashNil()
	case reflect.Bool:
		h.hashBool(v.Bool())
	case reflect.Uint8:
		h.hashInt(v.Uint(), 1)
	case reflect.Int8:
		h.hashInt(uint64(v.Int()), 1)
	case reflect.Uint16:
		h.hashInt(v.Uint(), 2)
	case reflect.Int16:
		h.hashInt(uint64(v.Int()), 2)
	case reflect.Uint32:
		h.hashInt(v.Uint(), 4)
	case reflect.Int32:
		h.hashInt(uint64(v.Int()), 4)
	case reflect.Uint64:
		h.hashInt(v.Uint(), 8)
	case reflect.Int64:
		h.hashInt(uint64(v.Int()), 8)
	case reflect.Uint:
		h.hashInt(v.Uint(), bits.UintSize>>3)
	case reflect.Int:
		h.hashInt(uint64(v.Int()), bits.UintSize>>3)
	case reflect.Float32:
		h.hashFloat(v.Float(), 4)
	case reflect.Float64:
		h.hashFloat(v.Float(), 8)
	case reflect.Complex64:
		h.hashComplex(v.Complex(), 4)
	case reflect.Complex128:
		h.hashComplex(v.Complex(), 8)
	case reflect.Uintptr:
		h.hashInt(v.Uint(), 8)
	case reflect.UnsafePointer:
		h.hashInt(uint64(v.Pointer()), 8)
	case reflect.String:
		h.hashString(v)
	case reflect.Func:
		h.visitFunc(v)
	case reflect.Chan:
		h.visitChan(v)
	case reflect.Interface:
		h.visitInterface(v)
	case reflect.Pointer:
		h.visitPointer(v)
	case reflect.Struct:
		h.visitStruct(v)
	case reflect.Array:
		h.visitArray(v)
	case reflect.Slice:
		h.visitSlice(v)
	case reflect.Map:
		h.visitMap(v)
	}
}

func (h *Hasher) visitFunc(v reflect.Value) {
	if h.addValue(v) {
		h.hashFunc(v)
	}
}

func (h *Hasher) visitChan(v reflect.Value) {
	if h.addValue(v) {
		h.hashChan(v)
	}
}

func (h *Hasher) visitInterface(v reflect.Value) {
	h.visit(v.Elem())
}

func (h *Hasher) visitPointer(v reflect.Value) {
	if v.IsNil() {
		h.hashNil()
		return
	}
	if !h.addValue(v) {
		return
	}
	elem := v.Elem()
	if h.addPointer(elem) {
		h.visit(elem)
	}
}

func (h *Hasher) visitStruct(v reflect.Value) {
	for _, field := range v.Fields() {
		h.idx++
		if h.addPointer(field) {
			h.visit(field)
		}
	}
}

func (h *Hasher) visitArray(v reflect.Value) {
	for i, n := 0, v.Len(); i < n; i++ {
		h.idx++
		elem := v.Index(i)
		if h.addPointer(elem) {
			h.visit(elem)
		}
	}
}

func (h *Hasher) visitSlice(v reflect.Value) {
	if v.IsNil() {
		h.hashNil()
		return
	}
	key := newSliceKey(v)
	if id, exists := h.lookupSlice(key); exists {
		h.hashIndex(id)
		return
	}
	if h.sliceValues == nil {
		h.sliceValues = make(map[sliceKey]int)
	}
	h.sliceValues[key] = h.idx
	backing := h.sliceBackings[key]
	if backing.group == 0 {
		backing.group = key.Ptr
	}
	groupId, exists := h.backingIds[backing.group]
	if !exists {
		groupId = len(h.backingIds)
		if h.backingIds == nil {
			h.backingIds = make(map[uintptr]int)
		}
		h.backingIds[backing.group] = groupId
	}
	h.hashIndex(groupId)
	h.hashIndex(int(backing.offset))
	h.hashIndex(v.Len())
	h.hashIndex(v.Cap())
	h.visitArray(v)
}

func (h *Hasher) visitMap(v reflect.Value) {
	if v.IsNil() {
		h.hashNil()
		return
	}
	if !h.addValue(v) {
		return
	}
	size := v.Len()
	h.hashIndex(size)
	entries := make(mapEntryList, 0, size)
	branch := h.clone()
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		branch.resetBranch(h.idx)
		branch.Add(key.Interface())
		branch.Add(value.Interface())
		entries = append(entries, mapEntry{
			key:   key,
			value: value,
			hash:  branch.Hash(),
			path:  h.mapEntryPath(key, h.orders),
		})
	}
	entries.Sort()
	for i := range size {
		h.visit(entries[i].key)
		h.visit(entries[i].value)
	}
}

func (h *Hasher) mapEntryPath(v reflect.Value, orders map[valueKey]string) string {
	switch v.Kind() {
	case reflect.Interface:
		if !v.IsNil() {
			return h.mapEntryPath(v.Elem(), orders)
		}
	case reflect.Chan, reflect.Pointer, reflect.UnsafePointer:
		return orders[valueKey{Addr: Address(v), Kind: 1}]
	}
	return ""
}

func (h *Hasher) prepareSliceBackings(occurrences []sliceOccurrence) {
	slices.SortFunc(occurrences, func(left, right sliceOccurrence) int {
		if left.typeName < right.typeName {
			return -1
		}
		if left.typeName > right.typeName {
			return 1
		}
		if left.key.Ptr < right.key.Ptr {
			return -1
		}
		if left.key.Ptr > right.key.Ptr {
			return 1
		}
		if left.path < right.path {
			return -1
		}
		if left.path > right.path {
			return 1
		}
		return 0
	})
	for start := 0; start < len(occurrences); {
		end := start + 1
		base := occurrences[start].key.Ptr
		rangeEnd := occurrences[start].key.capacityEnd()
		for end < len(occurrences) {
			current := occurrences[end].key
			previous := occurrences[end-1].key
			if current.ElemType != previous.ElemType || current.Ptr > rangeEnd || (current.Ptr == rangeEnd && current.Ptr != previous.Ptr) {
				break
			}
			rangeEnd = max(rangeEnd, current.capacityEnd())
			end++
		}
		for _, occurrence := range occurrences[start:end] {
			key := occurrence.key
			h.sliceBackings[key] = sliceBacking{
				group:  base,
				offset: key.Ptr - base,
			}
		}
		start = end
	}
}

func (h *Hasher) clone() *Hasher {
	return &Hasher{
		base:  h,
		mhash: &maphash.Hash{},
		idx:   h.idx,
	}
}

func (h *Hasher) resetBranch(idx int) {
	h.idx = idx
	h.dirty = false
	h.mhash.Reset()
	h.metadata.reset()
	clear(h.roots)
	h.roots = h.roots[:0]
	clear(h.values)
	clear(h.sliceValues)
	clear(h.sliceBackings)
	clear(h.backingIds)
	clear(h.orders)
}

func (h *Hasher) lookupValue(key valueKey) (int, bool) {
	for current := h; current != nil; current = current.base {
		if id, exists := current.values[key]; exists {
			return id, true
		}
	}
	return 0, false
}

func (h *Hasher) lookupSlice(key sliceKey) (int, bool) {
	for current := h; current != nil; current = current.base {
		if id, exists := current.sliceValues[key]; exists {
			return id, true
		}
	}
	return 0, false
}

func (h *Hasher) hashIndex(idx int) {
	h.hashInt(uint64(idx), bits.UintSize>>3)
}

func (h *Hasher) hashType(v reflect.Value) {
	if !v.IsValid() {
		return
	}
	h.hashBytes(NameOf(v.Type()))
}

func (h *Hasher) hashNil() {
	h.mhash.WriteByte(0)
}

func (h *Hasher) hashBool(v bool) {
	if v {
		h.mhash.WriteByte(255)
	} else {
		h.mhash.WriteByte(127)
	}
}

func (h *Hasher) hashInt(v uint64, bytes int) {
	for i := range bytes {
		h.mhash.WriteByte(byte(v >> (i << 3)))
	}
}

func (h *Hasher) hashFloat(v float64, bytes int) {
	if bytes == 4 {
		h.hashInt(uint64(math.Float32bits(float32(v))), bytes)
	} else {
		h.hashInt(math.Float64bits(v), bytes)
	}
}

func (h *Hasher) hashComplex(v complex128, bytes int) {
	h.hashFloat(real(v), bytes)
	h.hashFloat(imag(v), bytes)
}

func (h *Hasher) hashString(v reflect.Value) {
	h.hashBytes(v.String())
}

func (h *Hasher) hashFunc(v reflect.Value) {
	h.hashBytes(FuncNameOf(v))
}

func (h *Hasher) hashBytes(value string) {
	h.hashIndex(len(value))
	h.mhash.WriteString(value)
}

func (h *Hasher) hashChan(v reflect.Value) {
	h.hashIndex(v.Len())
	h.hashIndex(v.Cap())
}

func (h *Hasher) addValue(v reflect.Value) bool {
	return h.addNode(valueKey{
		Addr: Address(v),
		Kind: 1,
	})
}

func (h *Hasher) addPointer(v reflect.Value) bool {
	return h.addNode(valueKey{
		Addr: Addr{
			Ptr:  PtrOf(v),
			Type: v.Type(),
		},
		Kind: 0,
	})
}

func (h *Hasher) addNode(key valueKey) bool {
	id, exists := h.lookupValue(key)
	if h.values == nil {
		h.values = make(map[valueKey]int)
	}
	h.values[key] = h.idx
	if exists {
		h.hashIndex(id)
		return false
	}
	return true
}
