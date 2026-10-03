package reflex

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func TestSliceRelation(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	a := &[6]int{1, 2, 3, 4, 5, 6}
	s1 := strings.Clone("abcdef")
	s2 := strings.Clone("abcdef")
	s3 := []byte("abcdef")
	s4 := unsafe.String(unsafe.SliceData(s3), len(s3))
	items := []struct {
		slice1   any
		slice2   any
		expected SliceRelation
	}{
		// #1
		{
			s,
			[]byte{1, 2, 3, 4, 5, 6},
			SliceRelationNone,
		},
		// #2
		{
			s,
			[]int{1, 2, 3, 4, 5, 6},
			SliceRelationNone,
		},
		// #3
		{
			s,
			s,
			SliceRelationSelf,
		},
		// #4
		{
			s,
			s[0:6:6],
			SliceRelationSelf,
		},
		// #5
		{
			s,
			s[1:3],
			SliceRelationParent,
		},
		// #6
		{
			s[1:3],
			s,
			SliceRelationChild,
		},
		// #7
		{
			s[1:5],
			s[3:6],
			SliceRelationRelative,
		},
		// #8
		{
			s[3:6],
			s[1:5],
			SliceRelationRelative,
		},
		// #9
		{
			s[0:2],
			s[4:6],
			SliceRelationRelative,
		},
		// #10
		{
			s[4:6],
			s[0:2],
			SliceRelationRelative,
		},
		// #11
		{
			s[4:6],
			s[0:2:3],
			SliceRelationNone,
		},
		// #12
		{
			*a,
			[6]byte{1, 2, 3, 4, 5, 6},
			SliceRelationNone,
		},
		// #13
		{
			a,
			a,
			SliceRelationSelf,
		},
		// #14
		{
			a,
			a[0:6:6],
			SliceRelationSelf,
		},
		// #15
		{
			a,
			a[1:3],
			SliceRelationParent,
		},
		// #16
		{
			a[1:3],
			a,
			SliceRelationChild,
		},
		// #17
		{
			a[1:5],
			a[3:6],
			SliceRelationRelative,
		},
		// #18
		{
			a[3:6],
			a[1:5],
			SliceRelationRelative,
		},
		// #19
		{
			a[0:2],
			a[4:6],
			SliceRelationRelative,
		},
		// #20
		{
			a[4:6],
			a[0:2],
			SliceRelationRelative,
		},
		// #21
		{
			a[4:6],
			a[0:2:3],
			SliceRelationNone,
		},
		// #22
		{
			s[0:0],
			s[0:0],
			SliceRelationSelf,
		},
		// #23
		{
			s[0:0],
			s[:],
			SliceRelationChild,
		},
		// #24
		{
			s[:],
			s[0:0],
			SliceRelationParent,
		},
		// #25
		{
			s1,
			s2,
			SliceRelationNone,
		},
		// #26
		{
			s1,
			s1[0:6],
			SliceRelationSelf,
		},
		// #27
		{
			s1,
			s1[0:5],
			SliceRelationParent,
		},
		// #28
		{
			s1[0:5],
			s1,
			SliceRelationChild,
		},
		// #29
		{
			s1,
			s1[2:4],
			SliceRelationParent,
		},
		// #30
		{
			s1[2:4],
			s1,
			SliceRelationChild,
		},
		// #31
		{
			s1[2:4],
			s1[0:3],
			SliceRelationRelative,
		},
		// #32
		{
			s1[0:2],
			s1[1:4],
			SliceRelationRelative,
		},
		// #33
		{
			s4[0:3],
			s3[2:5],
			SliceRelationRelative,
		},
		// #34
		{
			s4[2:5],
			s3[0:3],
			SliceRelationRelative,
		},
	}
	for i, item := range items {
		s1 := NewSlice(reflect.ValueOf(item.slice1), 1)
		s2 := NewSlice(reflect.ValueOf(item.slice2), 2)
		actual := s1.Relation(s2)
		if actual != item.expected {
			t.Errorf("Test #%d failed: expected relation %s, got %s.", i+1, item.expected, actual)
		}
	}
}

func TestCommonParent(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	items := []struct {
		slice1 any
		s1, e1 int
		slice2 any
		s2, e2 int
	}{
		// #1
		{
			s[0:3], 0, 3,
			s[4:6], 4, 6,
		},
		// #2
		{
			s[4:6], 4, 6,
			s[0:3], 0, 3,
		},
		// #3
		{
			s[3:5], 3, 5,
			s[1:4], 1, 4,
		},
		// #4
		{
			s[1:3], 1, 3,
			s[4:6], 4, 6,
		},
		// #5
		{
			"abcd"[1:3], 1, 3,
			"abcd"[0:2], 0, 2,
		},
		// #6
		{
			"1234"[0:2], 0, 2,
			"1234"[1:4], 1, 4,
		},
	}
	for i, item := range items {
		s1 := NewSlice(reflect.ValueOf(item.slice1), 1)
		s2 := NewSlice(reflect.ValueOf(item.slice2), 2)
		r := s1.Relation(s2)
		if r != SliceRelationRelative {
			t.Errorf("Test #%d failed: expected relative relation between slices, got %s.", i+1, r)
			continue
		}
		parent := commonParent(s1, s2, -1)
		smin := min(item.s1, item.s2)
		a1 := parent.V.Slice(item.s1-smin, item.e1-smin)
		v1 := a1.Interface()
		if !reflect.DeepEqual(v1, item.slice1) {
			t.Errorf("Test #%d failed: expected first slice %v, got %v.", i+1, item.slice1, v1)
		}
		a2 := parent.V.Slice(item.s2-smin, item.e2-smin)
		v2 := a2.Interface()
		if !reflect.DeepEqual(v2, item.slice2) {
			t.Errorf("Test #%d failed: expected second slice %v, got %v.", i+1, item.slice2, v2)
		}
		if parent.Relation(NewSlice(a1, 1)) != SliceRelationParent {
			t.Errorf("Test #%d failed: the first slice is not child of a common parent.", i+1)
		}
		if parent.Relation(NewSlice(a2, 1)) != SliceRelationParent {
			t.Errorf("Test #%d failed: the second slice is not child of a common parent.", i+1)
		}
		if parent.Relation(s1) != SliceRelationParent {
			t.Errorf("Test #%d failed: the original first slice is not child of a common parent.", i+1)
		}
		if parent.Relation(s2) != SliceRelationParent {
			t.Errorf("Test #%d failed: the original second slice is not child of a common parent.", i+1)
		}
	}
}

func TestCommonParentStringAndByteSlice(t *testing.T) {
	backing := []byte("abcdef")
	str := unsafe.String(unsafe.SliceData(backing), len(backing))
	items := []struct {
		slice1     any
		s1, e1, k1 int
		slice2     any
		s2, e2, k2 int
	}{
		// #1
		{
			str[0:3], 0, 3, 3,
			backing[2:5], 2, 5, 6,
		},
		// #2
		{
			backing[2:5], 2, 5, 6,
			str[0:3], 0, 3, 3,
		},
	}
	for caseIndex, item := range items {
		s1 := NewSlice(reflect.ValueOf(item.slice1), 1)
		s2 := NewSlice(reflect.ValueOf(item.slice2), 2)
		if relation := s1.Relation(s2); relation != SliceRelationRelative {
			t.Errorf("Test #%d failed: expected relative relation between string and byte slice, got %s.", caseIndex+1, relation)
			continue
		}

		parent := commonParent(s1, s2, -1)
		if parent.V.Kind() != reflect.Slice || parent.V.Type().Elem() != reflect.TypeFor[byte]() {
			t.Errorf("Test #%d failed: expected a byte-slice common parent, got %s.", caseIndex+1, parent.V.Type())
			continue
		}

		minStart := min(item.s1, item.s2)
		for _, child := range []struct {
			slice                any
			start, end, capacity int
		}{
			{item.slice1, item.s1, item.e1, item.k1},
			{item.slice2, item.s2, item.e2, item.k2},
		} {
			start, end, capacity := parent.SliceOf(NewSlice(reflect.ValueOf(child.slice), 1))
			if start != child.start-minStart || end != child.end-minStart || capacity != child.capacity-minStart {
				t.Errorf("Test #%d failed: expected slice indices %d:%d:%d, got %d:%d:%d.",
					caseIndex+1, child.start-minStart, child.end-minStart, child.capacity-minStart, start, end, capacity)
				continue
			}
			expected := reflect.ValueOf(child.slice)
			if expected.Kind() == reflect.String {
				if got := string(parent.V.Slice(start, end).Bytes()); got != expected.String() {
					t.Errorf("Test #%d failed: common parent contains %q, expected %q.", caseIndex+1, got, expected.String())
				}
			} else if got := parent.V.Slice(start, end).Bytes(); !reflect.DeepEqual(got, expected.Interface()) {
				t.Errorf("Test #%d failed: common parent contains %v, expected %v.", caseIndex+1, got, expected.Interface())
			}
		}
	}
}

func TestSliceOf(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	items := []struct {
		parent  any
		child   any
		i, j, k int
	}{
		// #1
		{
			s,
			s[3:5:5],
			3, 5, 5,
		},
		// #2
		{
			s,
			s[0:3],
			0, 3, 6,
		},
		// #3
		{
			s[2:5:5],
			s[3:4:4],
			1, 2, 2,
		},
		// #4
		{
			"12345"[2:5],
			"12345"[3:4],
			1, 2, 2,
		},
	}
	for n, item := range items {
		parent := NewSlice(reflect.ValueOf(item.parent), 0)
		child := NewSlice(reflect.ValueOf(item.child), 1)
		r := parent.Relation(child)
		if r != SliceRelationParent {
			t.Errorf("Test #%d failed: expected parent relation between slices, got %s.", n+1, r)
			continue
		}
		i, j, k := parent.SliceOf(child)
		if i != item.i || j != item.j || k != item.k {
			t.Errorf("Test #%d failed: expected %d, %d, %d, got %d, %d, %d.", n+1, item.i, item.j, item.k, i, j, k)
		}
	}
}

func TestSliceMapStringAndByteSlice(t *testing.T) {
	backing := []byte("abcdef")
	str := unsafe.String(unsafe.SliceData(backing), len(backing))
	items := []struct {
		first, second any
	}{
		{str[0:3], backing[2:5]},
		{backing[2:5], str[0:3]},
	}
	for caseIndex, item := range items {
		m := NewSliceMap(-1)
		m.Add(reflect.ValueOf(item.first), 0)
		m.Add(reflect.ValueOf(item.second), 1)

		if len(m.parents) != 1 {
			t.Errorf("Test #%d failed: expected one common parent, got %d.", caseIndex+1, len(m.parents))
			continue
		}
		parent := m.parents[0]
		if !parent.IsVirtual() || parent.V.Kind() != reflect.Slice || parent.V.Type().Elem() != reflect.TypeFor[byte]() {
			t.Errorf("Test #%d failed: expected a virtual byte-slice parent, got id %d and type %s.", caseIndex+1, parent.Id, parent.V.Type())
			continue
		}
		if len(parent.Childs) != 2 {
			t.Errorf("Test #%d failed: expected two children, got %d.", caseIndex+1, len(parent.Childs))
			continue
		}
		if parent.Childs[0].Id != 0 || parent.Childs[1].Id != 1 {
			t.Errorf("Test #%d failed: expected children with IDs 0 and 1, got %d and %d.",
				caseIndex+1, parent.Childs[0].Id, parent.Childs[1].Id)
		}
	}
}

func TestSliceMapStringAndByteArray(t *testing.T) {
	backing := [6]byte{'a', 'b', 'c', 'd', 'e', 'f'}
	str := unsafe.String(&backing[0], len(backing))
	items := []struct {
		first, second any
		parentID      int
		childID       int
	}{
		{str[1:4], &backing, 1, 0},
		{&backing, str[1:4], 0, 1},
	}
	for caseIndex, item := range items {
		m := NewSliceMap(-1)
		m.Add(reflect.ValueOf(item.first), 0)
		m.Add(reflect.ValueOf(item.second), 1)

		if len(m.parents) != 1 {
			t.Errorf("Test #%d failed: expected one shared parent, got %d.", caseIndex+1, len(m.parents))
			continue
		}
		parent := m.parents[0]
		if parent.Id != item.parentID || parent.V.Kind() != reflect.Array || parent.V.Type().Elem() != reflect.TypeFor[byte]() {
			t.Errorf("Test #%d failed: expected byte-array parent with ID %d, got ID %d and type %s.",
				caseIndex+1, item.parentID, parent.Id, parent.V.Type())
			continue
		}
		if len(parent.Childs) != 1 || parent.Childs[0].Id != item.childID || parent.Childs[0].Parent != parent {
			t.Errorf("Test #%d failed: expected string child with ID %d.", caseIndex+1, item.childID)
		}
	}
}

func TestSliceMap(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5, 6}
	ss1 := s1[0:5]
	s2 := []int{1, 2, 3, 4, 5, 6}
	s3 := []byte{1, 2, 3, 4, 5, 6}
	s4 := strings.Clone("12345")
	s5 := strings.Clone("12345")

	m := NewSliceMap(-1)
	m.Add(reflect.ValueOf(s1[4:6]), 0)
	m.Add(reflect.ValueOf(s1[0:3]), 1)
	m.Add(reflect.ValueOf(s2[1:3]), 2)
	m.Add(reflect.ValueOf(ss1[1:4]), 3)
	m.Add(reflect.ValueOf(s2), 4)
	m.Add(reflect.ValueOf(s1), 5)
	m.Add(reflect.ValueOf(ss1[3:5]), 6)
	m.Add(reflect.ValueOf(s2[0:2]), 7)
	m.Add(reflect.ValueOf(s1), 8)
	m.Add(reflect.ValueOf(s3[4:6]), 9)
	m.Add(reflect.ValueOf(s3[1:3]), 10)
	m.Add(reflect.ValueOf(s3[0:2]), 11)
	m.Add(reflect.ValueOf(s3), 12)
	m.Add(reflect.ValueOf(s4[0:3]), 13)
	m.Add(reflect.ValueOf(s4[2:5]), 14)
	m.Add(reflect.ValueOf(s5[3:4]), 15)
	m.Add(reflect.ValueOf(s4), 16)
	m.Add(reflect.ValueOf(s5), 17)
	m.Add(reflect.ValueOf(s4), 18)

	res := []struct {
		parentId int
		childs   []int
	}{
		{
			5, []int{0, 1, 3, 6, 8},
		},
		{
			4, []int{2, 7},
		},
		{
			12, []int{9, 10, 11},
		},
		{
			16, []int{13, 14},
		},
		{
			17, []int{15},
		},
	}
	for i, p := range m.parents {
		if p.Id != res[i].parentId {
			t.Errorf("Parent #%d is wrong.", i+1)
			continue
		}
		for j, child := range p.Childs {
			if child.Id != res[i].childs[j] {
				t.Errorf("Child #%d of parent #%d is wrong.", j+1, i+1)
			}
		}
	}
}
