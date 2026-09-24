package reflex

import (
	"reflect"
	"testing"
)

func TestSliceRelation(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	a := &[6]int{1, 2, 3, 4, 5, 6}
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
			t.Errorf("Test #%d failed: expected first slice #%v, got #%v.", i+1, item.slice1, v1)
		}
		a2 := parent.V.Slice(item.s2-smin, item.e2-smin)
		v2 := a2.Interface()
		if !reflect.DeepEqual(v2, item.slice2) {
			t.Errorf("Test #%d failed: expected second slice #%v, got #%v.", i+1, item.slice2, v2)
		}
		if parent.Relation(NewSlice(a1, 1)) != SliceRelationParent {
			t.Errorf("Test #%d failed: the first slice is not child of a common parent.", i+1)
		}
		if parent.Relation(NewSlice(a2, 1)) != SliceRelationParent {
			t.Errorf("Test #%d failed: the second slice is not child of a common parent.", i+1)
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

func TestSliceMap(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5, 6}
	ss1 := s1[0:5]
	s2 := []int{1, 2, 3, 4, 5, 6}

	m := NewSliceMap(-1)
	m.Add(reflect.ValueOf(s1[4:6]), 0)
	m.Add(reflect.ValueOf(s1[0:3]), 1)
	m.Add(reflect.ValueOf(s2[1:3]), 2)
	m.Add(reflect.ValueOf(ss1[1:4]), 3)
	m.Add(reflect.ValueOf(s2), 4)
	m.Add(reflect.ValueOf(s1), 5)
	m.Add(reflect.ValueOf(ss1[3:5]), 6)
	m.Add(reflect.ValueOf(s2[0:2]), 7)

	res := []struct {
		parentId int
		childs   []int
	}{
		{
			5, []int{0, 1, 3, 6},
		},
		{
			4, []int{2, 7},
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
