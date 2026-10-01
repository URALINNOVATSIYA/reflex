package reflex

import (
	"math/bits"
	"strings"
	"testing"
)

type hasherPerson struct {
	Name string
	Age  int
}

type hasherUser struct {
	Name string
	Age  int
}

func TestHasher(t *testing.T) {
	items := []struct {
		value1 any
		value2 any
		equal  bool
	}{
		// #1
		{
			nil,
			nil,
			true,
		},
		// #2
		{
			false,
			false,
			true,
		},
		// #3
		{
			123,
			123,
			true,
		},
		// #4
		{
			3.14,
			3.14,
			true,
		},
		// #5
		{
			1.5 + 5.1i,
			1.5 + 5.1i,
			true,
		},
		// #6
		{
			PtrOf,
			PtrOf,
			true,
		},
		// #7
		{
			func() {},
			func() {},
			false,
		},
		// #8
		{
			make(chan<- bool, 5),
			make(chan<- bool, 5),
			true,
		},
		// #9
		{
			make(chan<- bool, 5),
			make(<-chan bool, 5),
			false,
		},
		// #10
		{
			"abc",
			"abc",
			true,
		},
		// #11
		{
			[3]byte{1, 2, 3},
			[3]byte{1, 2, 3},
			true,
		},
		// #12
		{
			[3]byte{1, 2, 3},
			[]byte{1, 2, 3},
			false,
		},
		// #13
		{
			func() any {
				s := []byte{1, 2, 3}
				return [2][]byte{s[0:2], s[1:3]}
			}(),
			func() any {
				s := []byte{1, 2, 3}
				return [2][]byte{s[0:2], s[1:3]}
			}(),
			true,
		},
		// #14
		{
			func() any {
				s := []byte{1, 2, 3}
				return [2][]byte{s[0:2], s[1:3]}
			}(),
			func() any {
				s := []byte{1, 2, 3}
				return [2][]byte{s[0:2], {2, 3}}
			}(),
			false,
		},
		// #15
		{
			func() any {
				a := [2]any{}
				a[0] = byte(123)
				a[1] = &a[0]
				return a
			}(),
			func() any {
				a := [2]any{}
				a[0] = byte(123)
				a[1] = &a[0]
				return a
			}(),
			true,
		},
		// #16
		{
			func() any {
				a := [2]any{}
				a[0] = byte(123)
				a[1] = &a[0]
				return a
			}(),
			func() any {
				a := [2]any{}
				b := byte(123)
				a[0] = b
				a[1] = &b
				return a
			}(),
			false,
		},
		// #17
		{
			func() any {
				a := [2]any{}
				a[0] = &a[1]
				a[1] = byte(123)
				return a
			}(),
			func() any {
				a := [2]any{}
				a[0] = &a[1]
				a[1] = byte(123)
				return a
			}(),
			true,
		},
		// #18
		{
			func() any {
				a := [2]any{}
				b := byte(123)
				a[0] = &b
				a[1] = b
				return a
			}(),
			func() any {
				a := [2]any{}
				a[0] = &a[1]
				a[1] = byte(123)
				return a
			}(),
			false,
		},
		// #19
		{
			func() any {
				b1 := byte(1)
				b2 := byte(2)
				return map[*byte]byte{&b1: 1, &b2: 2}
			}(),
			func() any {
				b1 := byte(1)
				b2 := byte(2)
				return map[*byte]byte{&b2: 2, &b1: 1}
			}(),
			true,
		},
		// #20
		{
			func() any {
				b1 := byte(1)
				b2 := byte(2)
				return map[*byte]byte{&b1: 1, &b2: 2}
			}(),
			func() any {
				b1 := byte(1)
				b2 := byte(1)
				return map[*byte]byte{&b2: 2, &b1: 1}
			}(),
			false,
		},
		// #21
		{
			hasherPerson{Name: "Kira", Age: 33},
			hasherPerson{Name: "Kira", Age: 33},
			true,
		},
		// #22
		{
			hasherPerson{Name: "Kira", Age: 33},
			hasherPerson{Name: "Kira", Age: 34},
			false,
		},
		// #23
		{
			hasherPerson{Name: "Kira", Age: 36},
			hasherUser{Name: "Kira", Age: 36},
			false,
		},
		// #24
		{
			func() any {
				a := [2]any{}
				p := &a
				a[0] = p
				a[1] = &a[0]
				return p
			}(),
			func() any {
				a := [2]any{}
				p := &a
				a[0] = p
				a[1] = &p
				return p
			}(),
			false,
		},
		// #25
		{
			func() any {
				m := map[any]any{}
				p := &m
				m[p] = m
				m[&p] = p
				return m
			}(),
			func() any {
				m := map[any]any{}
				p := &m
				m[&p] = p
				m[p] = m
				return m
			}(),
			true,
		},
		// #26
		{
			func() any {
				var x1, x2, x3 any
				x2 = &x3
				x1 = &x2
				x3 = &x1
				return x1
			}(),
			func() any {
				var x1, x2, x3 any
				x2 = &x3
				x1 = &x2
				x3 = &x2
				return x1
			}(),
			false,
		},
		// #27
		{
			struct {
				Value int `json:"value"`
			}{Value: 1},
			struct {
				Value int `json:"item"`
			}{Value: 1},
			false,
		},
		// #28
		{
			func() any {
				m := map[any]any{}
				p1 := &m
				p2 := &m
				m[&p1] = &p2
				m[&p2] = &p1
				return m
			}(),
			func() any {
				m := map[any]any{}
				p1 := &m
				p2 := &m
				m[&p2] = &p2
				m[&p1] = &p1
				return m
			}(),
			false,
		},
		// #29
		{
			func() any {
				return map[*int]string{
					new(int): "a",
					new(int): "b",
					new(int): "c",
					new(int): "d",
				}
			}(),
			func() any {
				return map[*int]string{
					new(int): "a",
					new(int): "b",
					new(int): "c",
					new(int): "d",
				}
			}(),
			true,
		},
		// #30
		{
			func() any {
				s := strings.Clone("abc")
				return [2]any{map[string]int{s: 1}, s}
			}(),
			func() any {
				s := strings.Clone("abc")
				return [2]any{map[string]int{s: 1}, s}
			}(),
			true,
		},
		// #31
		{
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [3]any{map[chan bool]int{ch1: 1, ch2: 2}, ch1, ch2}
			}(),
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [3]any{map[chan bool]int{ch1: 1, ch2: 2}, ch1, ch2}
			}(),
			true,
		},
		// #32
		{
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [3]any{map[chan bool]int{ch1: 1, ch2: 2}, ch1, ch2}
			}(),
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [3]any{map[chan bool]int{ch1: 2, ch2: 1}, ch1, ch2}
			}(),
			false,
		},
		// #33
		{
			func() any {
				s := strings.Clone("ab")
				return [2]string{s[:1], s}
			}(),
			func() any {
				s := strings.Clone("ax")
				return [2]string{s[:1], s}
			}(),
			false,
		},
		// #34
		{
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [4]any{
					map[chan bool]int{ch1: 0, ch2: 0},
					map[chan bool]int{ch2: 0, ch1: 0},
					map[chan bool]int{ch1: 0},
					map[chan bool]int{ch2: 0},
				}
			}(),
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [4]any{
					map[chan bool]int{ch1: 0, ch2: 0},
					map[chan bool]int{ch1: 0, ch2: 0},
					map[chan bool]int{ch1: 0},
					map[chan bool]int{ch2: 0},
				}
			}(),
			true,
		},
		// #35
		{
			func() any {
				s := []int{1, 2, 3, 4}
				return [2][]int{s[:1], s[1:2]}
			}(),
			func() any {
				s := []int{1, 2, 3, 4}
				return [2][]int{s[:1], s[1:2]}
			}(),
			true,
		},
		// #36
		{
			func() any {
				s := []int{1, 2, 3, 4}
				return [2][]int{s[:1], s[1:2]}
			}(),
			[2][]int{{1}, {2}},
			false,
		},
		// #37
		{
			func() any {
				first := make([]int, 1, 4)
				first[0] = 1
				second := make([]int, 1, 3)
				second[0] = 2
				return [2][]int{first, second}
			}(),
			func() any {
				first := make([]int, 1, 4)
				first[0] = 1
				second := make([]int, 1, 3)
				second[0] = 2
				return [2][]int{first, second}
			}(),
			true,
		},
		// #38
		{
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [4]any{
					map[any]any{ch1: 0, ch2: 0},
					map[any]any{ch2: 0, ch1: map[any]int{ch2: 0}},
					map[any]any{ch1: 0},
					map[any]any{ch2: map[any]any{ch1: map[any]any{ch1: 0}, ch2: 0}},
				}
			}(),
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [4]any{
					map[any]any{ch1: 0, ch2: 0},
					map[any]any{ch2: 0, ch1: map[any]int{ch2: 0}},
					map[any]any{ch1: 0},
					map[any]any{ch2: map[any]any{ch1: map[any]any{ch1: 0}, ch2: 0}},
				}
			}(),
			true,
		},
		// #39
		{
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [4]any{
					map[any]any{ch1: 0, ch2: 0},
					map[any]any{ch2: 0, ch1: map[any]int{ch2: 0}},
					map[any]any{ch1: 0},
					map[any]any{ch2: map[any]any{ch1: map[any]any{ch2: 0}, ch2: 0}},
				}
			}(),
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				return [4]any{
					map[any]any{ch1: 0, ch2: 0},
					map[any]any{ch2: 0, ch1: map[any]int{ch2: 0}},
					map[any]any{ch1: 0},
					map[any]any{ch2: map[any]any{ch1: map[any]any{ch1: 0}, ch2: 0}},
				}
			}(),
			false,
		},
		// #40
		{
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				s := []byte{1, 2, 3}
				return [4]any{
					map[any]any{ch1: 0, ch2: 0},
					map[any]any{ch2: s[1:2], ch1: map[any]any{ch2: []byte{1}}},
					map[any]any{ch1: s[1:2]},
					map[any]any{ch2: map[any]any{ch1: map[any]any{ch2: 0}, ch2: s[:1]}},
				}
			}(),
			func() any {
				ch1 := make(chan bool, 1)
				ch2 := make(chan bool, 1)
				s := []byte{1, 2, 3}
				return [4]any{
					map[any]any{ch1: 0, ch2: 0},
					map[any]any{ch2: s[1:2], ch1: map[any]any{ch2: []byte{1}}},
					map[any]any{ch1: s[1:2]},
					map[any]any{ch2: map[any]any{ch1: map[any]any{ch2: 0}, ch2: s[:1]}},
				}
			}(),
			true,
		},
		// #41
		{
			func() any {
				a := &[1]any{}
				a[0] = a
				return &a[0]
			}(),
			func() any {
				a1 := &[1]any{}
				a2 := &[1]any{}
				a2[0] = a1
				a1[0] = a2
				return &a1[0]
			}(),
			false,
		},
		// #42
		{
			func() any {
				first := []int{1, 2}
				second := []int{3, 4}
				nested := map[int][]int{
					1: first[:1],
					2: first[1:],
					3: second[:1],
					4: second[1:],
				}
				return map[*map[int][]int]bool{&nested: true}
			}(),
			func() any {
				first := []int{1, 4}
				second := []int{3, 2}
				nested := map[int][]int{
					1: first[:1],
					4: first[1:],
					3: second[:1],
					2: second[1:],
				}
				return map[*map[int][]int]bool{&nested: true}
			}(),
			false,
		},
	}
	hasher := NewHasher()
	for i, item := range items {
		hasher.Add(item.value1)
		hash1 := hasher.Hash()
		hasher.Reset()
		hasher.Add(item.value2)
		hash2 := hasher.Hash()
		hasher.Reset()
		if item.equal != (hash1 == hash2) {
			t.Errorf("Test #%d must return %t, got %t (hashes %x and %x).", i+1, item.equal, !item.equal, hash1, hash2)
		}
	}
}

func TestHasherTracksSliceAliasesAcrossRoots(t *testing.T) {
	hash := func(first, second []int) uint64 {
		hasher := NewHasher()
		hasher.Add(first)
		hasher.Add(second)
		return hasher.Hash()
	}

	backing := []int{1, 2, 3, 4}
	shared := hash(backing[:1], backing[1:2])

	first := make([]int, 1, 4)
	first[0] = 1
	second := make([]int, 1, 3)
	second[0] = 2
	separate := hash(first, second)
	if shared == separate {
		t.Fatalf("shared and separate backing arrays have identical hash %x", shared)
	}
}

func TestHasherFramesStringFieldBoundaries(t *testing.T) {
	secondFieldHeader := make([]byte, bits.UintSize>>3)
	for i := range secondFieldHeader {
		secondFieldHeader[i] = byte(uint64(5) >> (i * 8))
	}
	content := string(append(secondFieldHeader, "string"...))
	type pair struct {
		First  string
		Second string
	}
	first := pair{First: "", Second: content}
	second := pair{First: content, Second: ""}
	if Hash(first) == Hash(second) {
		t.Fatal("different string field values have identical hash")
	}
}

var hasherBenchmarkSink uint64

func BenchmarkHasherScalar(b *testing.B) {
	hasher := NewHasher()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		hasher.Reset()
		hasher.Add(123)
		hasherBenchmarkSink = hasher.Hash()
	}
}

func BenchmarkHasherStruct(b *testing.B) {
	value := hasherPerson{Name: "Kira", Age: 33}
	hasher := NewHasher()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		hasher.Reset()
		hasher.Add(value)
		hasherBenchmarkSink = hasher.Hash()
	}
}

func BenchmarkHasherSliceAliases(b *testing.B) {
	backing := []int{1, 2, 3, 4, 5, 6, 7, 8}
	value := [4][]int{backing[:3], backing[1:5], backing[3:7], backing[5:8]}
	hasher := NewHasher()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		hasher.Reset()
		hasher.Add(value)
		hasherBenchmarkSink = hasher.Hash()
	}
}

func BenchmarkHasherMap(b *testing.B) {
	ch1 := make(chan bool, 1)
	ch2 := make(chan bool, 1)
	value := [4]any{
		map[any]any{ch1: 0, ch2: 0},
		map[any]any{ch2: 0, ch1: map[any]int{ch2: 0}},
		map[any]any{ch1: 0},
		map[any]any{ch2: map[any]any{ch1: map[any]any{ch2: 0}, ch2: 0}},
	}
	hasher := NewHasher()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		hasher.Reset()
		hasher.Add(value)
		hasherBenchmarkSink = hasher.Hash()
	}
}
