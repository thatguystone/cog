package assert

import (
	"errors"
	"testing"
)

func TestT(t *testing.T) {
	a := A{t}

	a.T()
}

func TestBool(t *testing.T) {
	a := A{t}

	a.True(true, "expect true")
	a.MustTrue(true, "expect true")

	a.False(false, "expect false")
	a.MustFalse(false, "expect false")
}

func TestEqual(t *testing.T) {
	a := A{t}

	type table struct {
		e  interface{}
		g  interface{}
		ok bool
	}

	tests := []table{
		table{
			e:  1,
			g:  1,
			ok: true,
		},
		table{
			e:  1,
			g:  2,
			ok: false,
		},
		table{
			e:  "some long string",
			g:  "another long string",
			ok: false,
		},
		table{
			e:  1.0,
			g:  2,
			ok: false,
		},
	}

	f := 0.1
	sum := 0.0
	for i := 0; i < 10; i++ {
		sum += f
	}

	tests = append(tests, table{
		e:  float32(1.0),
		g:  float32(sum),
		ok: true,
	})

	tests = append(tests, table{
		e:  float64(1.0),
		g:  float64(sum),
		ok: true,
	})

	for _, test := range tests {
		if a.equal(test.e, test.g) != test.ok {
			eq := "=="
			if !test.ok {
				eq = "!="
			}

			t.Errorf("expected %#v %s %#v", test.e, eq, test.g)
		}
	}

	a.Equal(1, 1, "expect equal")
	a.MustEqual(1, 1, "expect equal")

	a.NotEqual(1, 2, "expect not equal")
	a.MustNotEqual(1, 2, "expect not equal")
}

func TestLen(t *testing.T) {
	a := A{t}

	a.Len([]int{1, 2}, 2, "expect length")
	a.MustLen([]int{1, 2}, 2, "expect length")

	a.LenNot([]int{1, 2}, 1, "expect no length")
	a.MustLenNot([]int{1, 2}, 1, "expect no length")
}

func TestContains(t *testing.T) {
	a := A{t}

	type table struct {
		c     interface{}
		v     interface{}
		found bool
		ok    bool
	}

	tests := []table{
		table{
			c:  1,
			ok: false,
		},
		table{
			c:  map[string]int{"test": 1, "test2": 2},
			ok: false,
		},
		table{
			c:     "some string",
			v:     "me st",
			ok:    true,
			found: true,
		},
		table{
			c:     []int{1, 2},
			v:     1,
			ok:    true,
			found: true,
		},
		table{
			c:     []int{1, 2},
			v:     3,
			ok:    true,
			found: false,
		},
		table{
			c:     []string{"test", "test2"},
			v:     "test2",
			ok:    true,
			found: true,
		},
		table{
			c:     []string{"test", "test2"},
			v:     "test3",
			ok:    true,
			found: false,
		},
	}

	for _, test := range tests {
		found, ok := a.contains(test.c, test.v)

		if ok != test.ok {
			negate := ""
			if !test.ok {
				negate = "not "
			}

			t.Errorf("expected %#v to %sbe iterable", test.c, negate)
		} else if found != test.found {
			in := "in"
			if !test.found {
				in = "not in"
			}

			t.Errorf("expected %#v %s %#v", test.c, in, test.v)
		}
	}

	a.Contains("test", "es", "expect to contain")
	a.Contains([]int{1, 2}, 1, "expect to contain")
	a.MustContain("test", "es", "expect to contain")

	a.NotContains("hi", "es", "expect to contain")
	a.NotContains([]int{1, 2}, 3, "expect to contain")
	a.MustNotContain("hi", "es", "expect to contain")
}

func TestIs(t *testing.T) {
	a := A{t}

	a.Is(1, 2, "expect same types")
	a.MustBe(1, 2, "expect same types")

	a.IsNot(1, a, "expect different types")
	a.IsNot(1, 2.0, "expect different types")
	a.MustNotBe(1, a, "expect different types")
}

func TestError(t *testing.T) {
	a := A{t}

	a.Error(errors.New("test"), "expect to get error")
	a.MustError(errors.New("test"), "expect to get error")

	a.NotError(nil, "expect no error")
	a.MustNotError(nil, "expect no error")
}

func TestPanic(t *testing.T) {
	a := A{t}

	panics := func() {
		panic("oh noez!")
	}

	a.Panic(panics, "expect panic")
	a.MustPanic(panics, "expect panic")

	a.NotPanic(func() {}, "expect no panic")
	a.MustNotPanic(func() {}, "expect no panic")
}

func ExampleA() {
	a := A{&testing.T{}}

	// These are just a few of the provided functions. Check out the full
	// documentation for everything.

	a.Equal(1, 1, "the universe is falling apart")
	a.NotEqual(1, 2, "those can't be equal!")

	panics := func() {
		panic("i get nervous sometimes")
	}
	a.Panic(panics, "this should always panic")

	// Get the original *testing.T
	a.T()

	// Output:
}
