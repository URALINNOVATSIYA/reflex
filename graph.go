package reflex

/*import (
	"encoding/binary"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type graphNode struct {
	label string
	edges []graphEdge
}

type graphEdge struct {
	label  string
	target int
}

type graph struct {
	nodes []graphNode
}

func (g *graph) addNode(label string) int {
	nodeId := len(g.nodes)
	g.nodes = append(g.nodes, graphNode{label: label})
	return nodeId
}

func (g *graph) addEdge(from, to int, label string) {
	g.nodes[from].edges = append(g.nodes[from].edges, graphEdge{label: label, target: to})
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
	group  int
	offset uintptr
}

type sliceOccurrence struct {
	key      sliceKey
	typeName string
}

type slotKey struct {
	group  int
	offset uintptr
}

type coloredEdge struct {
	label string
	color int
}

type incidence struct {
	direction byte
	label     string
	color     int
}

type GraphBuilder struct {
	graph         graph
	atoms         map[string]int
	references    map[valueKey]int
	sliceNodes    map[sliceKey]int
	sliceBackings map[sliceKey]sliceBacking
	backingNodes  map[int]int
	slots         map[slotKey]int
	slotValues    map[slotKey]bool
}

func NewGraphBuilder(values []reflect.Value) *GraphBuilder {
	builder := &GraphBuilder{
		atoms:         make(map[string]int),
		references:    make(map[valueKey]int),
		sliceNodes:    make(map[sliceKey]int),
		sliceBackings: make(map[sliceKey]sliceBacking),
		backingNodes:  make(map[int]int),
		slots:         make(map[slotKey]int),
		slotValues:    make(map[slotKey]bool),
	}
	builder.prepareSliceBackings(values)
	builder.createRootNode(values)
	return builder
}

func (b *GraphBuilder) createRootNode(values []reflect.Value) {
	rootId := b.graph.addNode("root")
	for i, value := range values {
		b.graph.addEdge(rootId, b.valueId(value), "root:"+strconv.Itoa(i))
	}
}

func (b *GraphBuilder) prepareSliceBackings(roots []reflect.Value) {
	occurrences := make([]sliceOccurrence, 0)
	seen := make(map[sliceKey]bool)
	activeRefs := make(map[valueKey]bool)
	activeSlices := make(map[sliceKey]bool)
	for _, root := range roots {
		b.collectSlices(root, &occurrences, seen, activeRefs, activeSlices)
	}
	slices.SortFunc(occurrences, func(left, right sliceOccurrence) int {
		if left.typeName != right.typeName {
			return strings.Compare(left.typeName, right.typeName)
		}
		if left.key.Ptr < right.key.Ptr {
			return -1
		}
		if left.key.Ptr > right.key.Ptr {
			return 1
		}
		if left.key.Len < right.key.Len {
			return -1
		}
		if left.key.Len > right.key.Len {
			return 1
		}
		return left.key.Cap - right.key.Cap
	})
	groupId := 0
	for start := 0; start < len(occurrences); {
		end := start + 1
		base := occurrences[start].key.Ptr
		rangeEnd := occurrences[start].key.capacityEnd()
		for end < len(occurrences) {
			current := occurrences[end].key
			previous := occurrences[end-1].key
			if current.ElemType != previous.ElemType || current.Ptr > rangeEnd ||
				(current.Ptr == rangeEnd && current.Ptr != previous.Ptr) {
				break
			}
			rangeEnd = max(rangeEnd, current.capacityEnd())
			end++
		}
		for _, occurrence := range occurrences[start:end] {
			b.sliceBackings[occurrence.key] = sliceBacking{
				group:  groupId,
				offset: occurrence.key.Ptr - base,
			}
		}
		groupId++
		start = end
	}
}

func (b *GraphBuilder) collectSlices(
	v reflect.Value,
	occurrences *[]sliceOccurrence,
	seen map[sliceKey]bool,
	activeRefs map[valueKey]bool,
	activeSlices map[sliceKey]bool,
) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Interface:
		if !v.IsNil() {
			b.collectSlices(v.Elem(), occurrences, seen, activeRefs, activeSlices)
		}
	case reflect.Pointer:
		if v.IsNil() {
			return
		}
		key := valueKey{Addr: Address(v), Kind: 1}
		if activeRefs[key] {
			return
		}
		activeRefs[key] = true
		b.collectSlices(v.Elem(), occurrences, seen, activeRefs, activeSlices)
		delete(activeRefs, key)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			b.collectSlices(v.Field(i), occurrences, seen, activeRefs, activeSlices)
		}
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			b.collectSlices(v.Index(i), occurrences, seen, activeRefs, activeSlices)
		}
	case reflect.Slice:
		if v.IsNil() {
			return
		}
		key := newSliceKey(v)
		if !seen[key] {
			seen[key] = true
			*occurrences = append(*occurrences, sliceOccurrence{key: key, typeName: NameOf(v.Type().Elem())})
		}
		if activeSlices[key] {
			return
		}
		activeSlices[key] = true
		for i := 0; i < v.Len(); i++ {
			b.collectSlices(v.Index(i), occurrences, seen, activeRefs, activeSlices)
		}
		delete(activeSlices, key)
	case reflect.Map:
		if v.IsNil() {
			return
		}
		key := valueKey{Addr: Address(v), Kind: 2}
		if activeRefs[key] {
			return
		}
		activeRefs[key] = true
		iter := v.MapRange()
		for iter.Next() {
			b.collectSlices(iter.Key(), occurrences, seen, activeRefs, activeSlices)
			b.collectSlices(iter.Value(), occurrences, seen, activeRefs, activeSlices)
		}
		delete(activeRefs, key)
	}
}

func (b *GraphBuilder) valueId(v reflect.Value) int {
	if !v.IsValid() {
		return b.atom("nil")
	}
	if v.Kind() == reflect.Slice {
		return b.slice(v)
	}
	if label, ok := b.atomLabel(v); ok {
		return b.atom(label)
	}
	if key, ok := b.referenceKey(v); ok {
		if id, exists := b.references[key]; exists {
			return id
		}
		id := b.graph.addNode("")
		b.references[key] = id
		b.expand(id, v)
		return id
	}
	id := b.graph.addNode("")
	b.expand(id, v)
	return id
}

func (b *GraphBuilder) atom(label string) int {
	if id, exists := b.atoms[label]; exists {
		return id
	}
	id := b.graph.addNode(label)
	b.atoms[label] = id
	return id
}

func (b *GraphBuilder) atomLabel(v reflect.Value) (string, bool) {
	typeName := NameOf(v.Type())
	switch v.Kind() {
	case reflect.Bool:
		return typeName + ":" + strconv.FormatBool(v.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return typeName + ":" + strconv.FormatInt(v.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return typeName + ":" + strconv.FormatUint(v.Uint(), 10), true
	case reflect.Float32:
		return typeName + ":" + strconv.FormatUint(uint64(math.Float32bits(float32(v.Float()))), 16), true
	case reflect.Float64:
		return typeName + ":" + strconv.FormatUint(math.Float64bits(v.Float()), 16), true
	case reflect.Complex64:
		value := complex64(v.Complex())
		return typeName + ":" + strconv.FormatUint(uint64(math.Float32bits(real(value))), 16) + ":" + strconv.FormatUint(uint64(math.Float32bits(imag(value))), 16), true
	case reflect.Complex128:
		value := v.Complex()
		return typeName + ":" + strconv.FormatUint(math.Float64bits(real(value)), 16) + ":" + strconv.FormatUint(math.Float64bits(imag(value)), 16), true
	case reflect.String:
		return typeName + ":" + v.String(), true
	case reflect.UnsafePointer:
		return typeName + ":" + strconv.FormatUint(uint64(v.Pointer()), 16), true
	}
	return "", false
}

func (b *GraphBuilder) referenceKey(v reflect.Value) (valueKey, bool) {
	switch v.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Chan, reflect.Func:
		if !v.IsNil() {
			kind := byte(v.Kind())
			return valueKey{Addr: Address(v), Kind: kind}, true
		}
	}
	return valueKey{}, false
}

func (b *GraphBuilder) slice(v reflect.Value) int {
	key := newSliceKey(v)
	if id, exists := b.sliceNodes[key]; exists {
		return id
	}
	id := b.graph.addNode("")
	b.sliceNodes[key] = id
	if v.IsNil() {
		b.graph.nodes[id].label = NameOf(v.Type()) + ":nil"
		return id
	}
	backing, exists := b.sliceBackings[key]
	if !exists {
		backing = sliceBacking{group: -1}
	}
	group := b.backingNode(backing.group, v.Type().Elem())
	b.graph.nodes[id].label = NameOf(v.Type()) + ":" + strconv.Itoa(v.Len()) + ":" + strconv.Itoa(v.Cap()) + ":" + strconv.FormatUint(uint64(backing.offset), 10)
	b.graph.addEdge(id, group, "backing")
	for i := 0; i < v.Len(); i++ {
		offset := backing.offset + uintptr(i)*v.Type().Elem().Size()
		slotKey := slotKey{group: backing.group, offset: offset}
		slot, exists := b.slots[slotKey]
		if !exists {
			slot = b.graph.addNode("backing-slot")
			b.slots[slotKey] = slot
			b.graph.addEdge(group, slot, "slot:"+strconv.FormatUint(uint64(offset), 10))
		}
		b.graph.addEdge(id, slot, "element:"+strconv.Itoa(i))
		if !b.slotValues[slotKey] {
			b.slotValues[slotKey] = true
			b.graph.addEdge(slot, b.valueId(v.Index(i)), "value")
		}
	}
	return id
}

func (b *GraphBuilder) backingNode(group int, elemType reflect.Type) int {
	if id, exists := b.backingNodes[group]; exists {
		return id
	}
	id := b.graph.addNode("backing:" + NameOf(elemType))
	b.backingNodes[group] = id
	return id
}

func (b *GraphBuilder) expand(id int, v reflect.Value) {
	switch v.Kind() {
	case reflect.Interface:
		b.graph.nodes[id].label = NameOf(v.Type())
		if v.IsNil() {
			b.graph.nodes[id].label += ":nil"
		} else {
			b.graph.addEdge(id, b.valueId(v.Elem()), "dynamic")
		}
	case reflect.Pointer:
		b.graph.nodes[id].label = NameOf(v.Type())
		if v.IsNil() {
			b.graph.nodes[id].label += ":nil"
		} else {
			b.graph.addEdge(id, b.valueId(v.Elem()), "target")
		}
	case reflect.Struct:
		b.graph.nodes[id].label = NameOf(v.Type())
		for i := 0; i < v.NumField(); i++ {
			b.graph.addEdge(id, b.valueId(v.Field(i)), "field:"+strconv.Itoa(i))
		}
	case reflect.Array:
		b.graph.nodes[id].label = NameOf(v.Type())
		for i := 0; i < v.Len(); i++ {
			b.graph.addEdge(id, b.valueId(v.Index(i)), "element:"+strconv.Itoa(i))
		}
	case reflect.Map:
		b.graph.nodes[id].label = NameOf(v.Type())
		if v.IsNil() {
			b.graph.nodes[id].label += ":nil"
			return
		}
		iter := v.MapRange()
		for iter.Next() {
			entry := b.graph.addNode("map-entry")
			b.graph.addEdge(id, entry, "entry")
			b.graph.addEdge(entry, b.valueId(iter.Key()), "key")
			b.graph.addEdge(entry, b.valueId(iter.Value()), "value")
		}
	case reflect.Chan:
		b.graph.nodes[id].label = NameOf(v.Type())
		if v.IsNil() {
			b.graph.nodes[id].label += ":nil"
		} else {
			b.graph.nodes[id].label += ":" + strconv.Itoa(v.Len()) + ":" + strconv.Itoa(v.Cap())
		}
	case reflect.Func:
		b.graph.nodes[id].label = NameOf(v.Type()) + ":" + FuncNameOf(v)
	case reflect.Slice:
		panic("slice values must be built through slice")
	default:
		b.graph.nodes[id].label = NameOf(v.Type()) + ":unsupported"
	}
}

func (b *GraphBuilder) Encode() []byte {
	marks := make([]int, len(b.graph.nodes))
	return b.search(marks, 1)
}

func (b *GraphBuilder) search(marks []int, nextMark int) []byte {
	colors := b.refine(marks)
	cell := b.nonSingletonCell(colors)
	if len(cell) == 0 {
		return b.serialize(colors, marks)
	}
	var best []byte
	for _, node := range cell {
		individualized := slices.Clone(marks)
		individualized[node] = nextMark
		candidate := b.search(individualized, nextMark+1)
		if best == nil || slices.Compare(candidate, best) < 0 {
			best = candidate
		}
	}
	return best
}

func (b *GraphBuilder) refine(marks []int) []int {
	signatures := make([]string, len(b.graph.nodes))
	incidenceScratch := make([]incidence, 0)
	for i, node := range b.graph.nodes {
		var signature strings.Builder
		b.writeToken(&signature, node.label)
		b.writeUint(&signature, uint64(marks[i]))
		signatures[i] = signature.String()
	}
	colors := b.assignColors(signatures)
	incoming := make([][]coloredEdge, len(b.graph.nodes))
	for range len(b.graph.nodes) {
		for i := range incoming {
			incoming[i] = incoming[i][:0]
		}
		for source, node := range b.graph.nodes {
			for _, edge := range node.edges {
				incoming[edge.target] = append(incoming[edge.target], coloredEdge{
					label: edge.label,
					color: colors[source],
				})
			}
		}
		for i, node := range b.graph.nodes {
			edges := b.incidences(incidenceScratch[:0], node, colors, incoming[i])
			incidenceScratch = edges
			var signature strings.Builder
			b.writeUint(&signature, uint64(colors[i]))
			b.writeToken(&signature, node.label)
			b.writeUint(&signature, uint64(marks[i]))
			b.writeUint(&signature, uint64(len(edges)))
			for _, edge := range edges {
				signature.WriteByte(edge.direction)
				b.writeToken(&signature, edge.label)
				b.writeUint(&signature, uint64(edge.color))
			}
			signatures[i] = signature.String()
		}
		updated := b.assignColors(signatures)
		if slices.Equal(colors, updated) {
			return colors
		}
		colors = updated
	}
	return colors
}

func (b *GraphBuilder) incidences(incidences []incidence, node graphNode, colors []int, incoming []coloredEdge) []incidence {
	incidences = incidences[:0]
	incidences = slices.Grow(incidences, len(node.edges)+len(incoming))
	for _, edge := range node.edges {
		incidences = append(incidences, incidence{direction: 0, label: edge.label, color: colors[edge.target]})
	}
	for _, edge := range incoming {
		incidences = append(incidences, incidence{direction: 1, label: edge.label, color: edge.color})
	}
	slices.SortFunc(incidences, func(left, right incidence) int {
		if left.direction != right.direction {
			return int(left.direction) - int(right.direction)
		}
		if comparison := strings.Compare(left.label, right.label); comparison != 0 {
			return comparison
		}
		return left.color - right.color
	})
	return incidences
}

func (b *GraphBuilder) assignColors(signatures []string) []int {
	sorted := slices.Clone(signatures)
	slices.Sort(sorted)
	unique := sorted[:0]
	for _, signature := range sorted {
		if len(unique) == 0 || unique[len(unique)-1] != signature {
			unique = append(unique, signature)
		}
	}
	colors := make([]int, len(signatures))
	for i, signature := range signatures {
		colors[i], _ = slices.BinarySearch(unique, signature)
	}
	return colors
}

func (b *GraphBuilder) nonSingletonCell(colors []int) []int {
	counts := make(map[int]int)
	for _, color := range colors {
		counts[color]++
	}
	selected, selectedSize := -1, int(^uint(0)>>1)
	for color, size := range counts {
		if size > 1 && size < selectedSize {
			selected, selectedSize = color, size
		}
	}
	if selected < 0 {
		return nil
	}
	cell := make([]int, 0, selectedSize)
	for node, color := range colors {
		if color == selected {
			cell = append(cell, node)
		}
	}
	return cell
}

func (b *GraphBuilder) serialize(colors, marks []int) []byte {
	order := make([]int, len(b.graph.nodes))
	for node, color := range colors {
		order[color] = node
	}
	canonicalIDs := make([]int, len(b.graph.nodes))
	for canonicalID, node := range order {
		canonicalIDs[node] = canonicalID
	}
	encoded := binary.BigEndian.AppendUint64(nil, uint64(len(b.graph.nodes)))
	for _, nodeID := range order {
		node := b.graph.nodes[nodeID]
		encoded = b.appendToken(encoded, node.label)
		encoded = binary.BigEndian.AppendUint64(encoded, uint64(marks[nodeID]))
		edges := slices.Clone(node.edges)
		slices.SortFunc(edges, func(left, right graphEdge) int {
			if comparison := strings.Compare(left.label, right.label); comparison != 0 {
				return comparison
			}
			return canonicalIDs[left.target] - canonicalIDs[right.target]
		})
		encoded = binary.BigEndian.AppendUint64(encoded, uint64(len(edges)))
		for _, edge := range edges {
			encoded = b.appendToken(encoded, edge.label)
			encoded = binary.BigEndian.AppendUint64(encoded, uint64(canonicalIDs[edge.target]))
		}
	}
	return encoded
}

func (b *GraphBuilder) appendToken(dst []byte, value string) []byte {
	dst = binary.BigEndian.AppendUint64(dst, uint64(len(value)))
	return append(dst, value...)
}

func (b *GraphBuilder) writeToken(builder *strings.Builder, value string) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	builder.Write(size[:])
	builder.WriteString(value)
}

func (b *GraphBuilder) writeUint(builder *strings.Builder, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	builder.Write(encoded[:])
}
*/
