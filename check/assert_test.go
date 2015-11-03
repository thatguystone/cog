package check

import (
	"errors"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	c := New(t)

	c.Equal("", format())
	c.Equal("test", format("test"))
	c.Equal("test 1", format("test %d", 1))
}

func TestCallerInfo(t *testing.T) {
	c := New(t)

	c.Contains(callerInfo(), "assert_test.go")
	func() {
		c.Contains(callerInfo(), "assert_test.go")
	}()
}

func TestBool(t *testing.T) {
	c := New(t)

	c.True(true, "expect true")
	c.MustTrue(true, "expect true")

	c.False(false, "expect false")
	c.MustFalse(false, "expect false")
}

func TestEqual(t *testing.T) {
	c := New(t)

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
		table{
			e:  nil,
			g:  []byte(nil),
			ok: true,
		},
		table{
			e:  []byte(nil),
			g:  []byte(nil),
			ok: true,
		},
		table{
			e:  []byte(nil),
			g:  nil,
			ok: false,
		},
		table{
			e:  nil,
			g:  nil,
			ok: true,
		},
		table{
			e:  nil,
			g:  1,
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
		if c.equal(test.e, test.g) != test.ok {
			eq := "=="
			if !test.ok {
				eq = "!="
			}

			c.Errorf("expected %#v %s %#v", test.e, eq, test.g)
		}
	}

	c.Equal(1, 1, "expect equal")
	c.MustEqual(1, 1, "expect equal")

	c.NotEqual(1, 2, "expect not equal")
	c.MustNotEqual(1, 2, "expect not equal")

	// Interface ints vs untyped ints don't compare nicely
	v := uint64(1)
	c.MustEqual(1, v, "expect equal")
}

func TestLen(t *testing.T) {
	c := New(t)

	c.Len([]int{1, 2}, 2, "expect length")
	c.MustLen([]int{1, 2}, 2, "expect length")

	c.LenNot([]int{1, 2}, 1, "expect no length")
	c.MustLenNot([]int{1, 2}, 1, "expect no length")
}

func TestContains(t *testing.T) {
	c := New(t)

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
			c:     []byte("some string"),
			v:     []byte("me st"),
			ok:    true,
			found: true,
		},
		table{
			c:     []byte("some string"),
			v:     []byte("some string"),
			ok:    true,
			found: true,
		},
		table{
			c:     []byte("some string"),
			v:     []byte("no way"),
			ok:    true,
			found: false,
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
		found, ok := c.contains(test.c, test.v)

		if ok != test.ok {
			negate := ""
			if !test.ok {
				negate = "not "
			}

			c.Errorf("expected %#v to %sbe iterable", test.c, negate)
		} else if found != test.found {
			in := "in"
			if !test.found {
				in = "not in"
			}

			c.Errorf("expected %#v %s %#v", test.c, in, test.v)
		}
	}

	c.Contains("test", "es", "expect to contain")
	c.Contains([]int{1, 2}, 1, "expect to contain")
	c.MustContain("test", "es", "expect to contain")

	c.NotContains("hi", "es", "expect to contain")
	c.NotContains([]int{1, 2}, 3, "expect to contain")
	c.MustNotContain("hi", "es", "expect to contain")
}

func TestIs(t *testing.T) {
	c := New(t)

	c.Is(1, 2, "expect same types")
	c.MustBe(1, 2, "expect same types")

	c.IsNot(1, c, "expect different types")
	c.IsNot(1, 2.0, "expect different types")
	c.MustNotBe(1, c, "expect different types")
}

func TestError(t *testing.T) {
	c := New(t)

	c.Error(errors.New("test"), "expect to get error")
	c.MustError(errors.New("test"), "expect to get error")

	c.NotError(nil, "expect no error")
	c.MustNotError(nil, "expect no error")
}

func TestPanic(t *testing.T) {
	c := New(t)

	panics := func() {
		panic("oh noez!")
	}

	c.Panic(panics, "expect panic")
	c.MustPanic(panics, "expect panic")

	c.NotPanic(func() {}, "expect no panic")
	c.MustNotPanic(func() {}, "expect no panic")
}

func TestUntil(t *testing.T) {
	c := New(t)

	i := 0
	c.Until(time.Second, func() bool { i++; return i > 10 }, "failed")
}
