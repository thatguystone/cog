package check

import (
	"errors"
	"reflect"
	"testing"
)

func compareAny(a, b any) int {
	return compare(reflect.ValueOf(a), reflect.ValueOf(b))
}

func TestSortMap(t *testing.T) {
	vals := []any{
		nil,
		bool(false),
		bool(true),
		int(1),
		int(2),
		int(10),
		int8(-5),
		int8(5),
		int32(5),
		uint(10),
		uint8(5),
		uint16(1),
		uint64(0),
		uint64(100),
		float32(-1.0),
		float32(10.0),
		float64(-10.0),
		float64(10.0),
		complex64(1 + 3i),
		complex64(1 + 5i),
		complex128(-1 + 3i),
		complex128(0 + 3i),
		[3]int{0, 0, 0},
		[3]int{0, 0, 1},
		[3]int{0, 1, 1},
		[3]int{1, 1, 1},
		[3]int8{},
		[3]int16{},
		(chan int)(nil),
		make(chan int),
		make(chan int8),
		make(chan int16),
		make(chan string),
		make(chan struct{}),
		new(int),
		new(int8),
		new(int16),
		new(int32),
		errors.New("1"),
		errors.New("2"),
		"a",
		"b",
		"c",
		"d",
		struct{}{},
	}

	m := make(map[any]any)
	for _, val := range vals {
		m[val] = val
	}

	s := make([]any, 0, len(vals))
	for _, kv := range sortMap(reflect.ValueOf(m)) {
		s = append(s, kv.k.Interface())
	}

	Equal(t, s, vals)
}

func TestCompareEquals(t *testing.T) {
	var (
		st0 struct{}
		st1 struct{ a int }
		ch  = make(chan int)
	)

	Equal(t, compareAny(0, 0), 0)
	Equal(t, compareAny([3]int{}, [3]int{}), 0)
	Equal(t, compareAny(st0, st0), 0)
	Equal(t, compareAny(st1, st1), 0)
	Equal(t, compareAny(ch, ch), 0)
}

func TestCompareErrors(t *testing.T) {
	Panics(t, func() {
		compareAny([]int{}, []int{})
	})
}

func TestCompareTypes(t *testing.T) {
	type (
		testInt  int
		testInt2 int
	)

	cmpTypes := func(a, b any) int {
		return compareTypes(reflect.TypeOf(a), reflect.TypeOf(b))
	}

	Equal(t, cmpTypes(0, 0), 0)
	Equal(t, cmpTypes(0, "str"), -1)
	Equal(t, cmpTypes(0, testInt(0)), -1)
	Equal(t, cmpTypes(testInt(0), 0), +1)
	Equal(t, cmpTypes(testInt(0), testInt2(0)), -1)
	Equal(t, cmpTypes(testInt2(0), testInt(0)), +1)
}

func TestCompareBool(t *testing.T) {
	Equal(t, compareBool(false, false), 0)
	Equal(t, compareBool(true, true), 0)
	Equal(t, compareBool(false, true), -1)
	Equal(t, compareBool(true, false), +1)
}

func TestCompareNil(t *testing.T) {
	var (
		ptr    = new(any)
		nilPtr *any
	)

	tests := []struct {
		a  any
		b  any
		ok bool
		c  int
	}{
		{
			a:  nilPtr,
			b:  nilPtr,
			ok: true,
			c:  0,
		},
		{
			a:  ptr,
			b:  nilPtr,
			ok: true,
			c:  +1,
		},
		{
			a:  nilPtr,
			b:  ptr,
			ok: true,
			c:  -1,
		},
		{
			a:  ptr,
			b:  ptr,
			ok: false,
			c:  0,
		},
	}

	for _, test := range tests {
		c, ok := compareNil(reflect.ValueOf(test.a), reflect.ValueOf(test.b))
		Equal(t, ok, test.ok)
		Equal(t, c, test.c)
	}
}
