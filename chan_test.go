package reflex

import (
	"reflect"
	"testing"
)

func TestChanElems(t *testing.T) {
	items := []struct {
		ch    any
		elems []any
	}{
		{
			make(chan int),
			nil,
		},
		{
			make(chan int, 2),
			nil,
		},
		{
			func() any {
				ch := make(chan int, 4)
				ch <- 1
				ch <- 2
				ch <- 3
				<-ch
				<-ch
				ch <- 4
				ch <- 5
				ch <- 6
				return ch
			}(),
			[]any{3, 4, 5, 6},
		},
		{
			func() any {
				ch := make(chan struct{}, 2)
				ch <- struct{}{}
				ch <- struct{}{}
				return ch
			}(),
			[]any{struct{}{}, struct{}{}},
		},
	}
	for i, item := range items {
		elems := ChanElems(reflect.ValueOf(item.ch))
		if len(elems) != len(item.elems) {
			t.Fatalf("Test %d failed: ChanElems() returned %d items, expected %d", i+1, len(elems), len(item.elems))
		}
		for j, elem := range elems {
			if elem.Interface() != item.elems[j] {
				t.Errorf("Test %d failed: ChanElems()[%d] = %v, expected %v", i+1, j, elem.Interface(), item.elems[j])
			}
		}
	}
}

func TestChanElemIter(t *testing.T) {
	items := []struct {
		ch    any
		elems []any
	}{
		{
			make(chan int),
			nil,
		},
		{
			make(chan int, 2),
			nil,
		},
		{
			func() any {
				ch := make(chan int, 4)
				ch <- 1
				ch <- 2
				ch <- 3
				<-ch
				<-ch
				ch <- 4
				ch <- 5
				ch <- 6
				return ch
			}(),
			[]any{3, 4, 5, 6},
		},
		{
			func() any {
				ch := make(chan struct{}, 2)
				ch <- struct{}{}
				ch <- struct{}{}
				return ch
			}(),
			[]any{struct{}{}, struct{}{}},
		},
	}
	for i, item := range items {
		elems := ChanElemIter(reflect.ValueOf(item.ch))
		j := 0
		for elem := range elems {
			if elem.Interface() != item.elems[j] {
				t.Errorf("Test %d failed: ChanElemIter()[%d] = %v, expected %v", i+1, j, elem.Interface(), item.elems[j])
			}
			j++
		}
		if j != len(item.elems) {
			t.Fatalf("Test %d failed: ChanElemIter() returned %d items, expected %d", i+1, j, len(item.elems))
		}
	}
}

func TestChanElemIterStopOnFalse(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	ch <- 3
	count := 0
	for range ChanElemIter(reflect.ValueOf(ch)) {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("ChanElemIter yielded %d items before stopping, want 1", count)
	}
}
