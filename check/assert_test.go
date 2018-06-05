package check

import (
	"errors"
	"fmt"
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
	c.Must.True(true, "expect true")

	c.False(false, "expect false")
	c.Must.False(false, "expect false")
}

func TestEqual(t *testing.T) {
	c := New(t)

	type table struct {
		e     interface{}
		g     interface{}
		equal bool
	}

	tests := []table{
		table{
			e:     1,
			g:     1,
			equal: true,
		},
		table{
			e:     1,
			g:     2,
			equal: false,
		},
		table{
			e:     "some long string",
			g:     "another long string",
			equal: false,
		},
		table{
			e:     1.0,
			g:     2,
			equal: false,
		},
		table{
			e:     nil,
			g:     []byte(nil),
			equal: false,
		},
		table{
			e:     []byte(nil),
			g:     []byte(nil),
			equal: true,
		},
		table{
			e:     []byte(nil),
			g:     nil,
			equal: false,
		},
		table{
			e:     nil,
			g:     nil,
			equal: true,
		},
		table{
			e:     nil,
			g:     1,
			equal: false,
		},
	}

	f := 0.1
	sum := 0.0
	for i := 0; i < 10; i++ {
		sum += f
	}

	tests = append(tests, table{
		e:     float32(1.0),
		g:     float32(sum),
		equal: true,
	})

	tests = append(tests, table{
		e:     float64(1.0),
		g:     float64(sum),
		equal: true,
	})

	for i, test := range tests {
		test := test // Capture test
		c.Run(fmt.Sprintf("%d", i),
			func(c *C) {
				var ok bool

				if test.equal {
					ok = c.Equal(test.g, test.e)
				} else {
					ok = c.NotEqual(test.g, test.e)
				}

				c.True(ok)
			})
	}
}

func TestEqualExtras(t *testing.T) {
	c := New(t)

	c.Equal(1, 1, "expect equal")
	c.Must.Equal(1, 1, "expect equal")

	c.NotEqual(1, 2, "expect not equal")
	c.Must.NotEqual(1, 2, "expect not equal")

	// Interface ints vs untyped ints don't compare nicely
	v := uint64(1)
	c.Must.Equal(1, v, "expect equal")
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
			ok: true,
		},
		table{
			c:     map[string]int{"test": 1, "test2": 2},
			v:     "test",
			ok:    true,
			found: true,
		},
		table{
			c:     map[string]int{"test": 1, "test2": 2},
			v:     "test123",
			ok:    true,
			found: false,
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
			v:     []byte("some stringy stuff"),
			ok:    true,
			found: false,
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

	for i, test := range tests {
		c.Run(fmt.Sprintf("%d", i),
			func(c *C) {
				a := newNoopAssert(c.TB)

				found, ok := a.contains(test.c, test.v)

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

					c.Errorf("expected %#v %s %#v", test.v, in, test.c)
				}
			})
	}
}

func TestContainsExtra(t *testing.T) {
	c := New(t)

	c.Contains("test", "es", "expect to contain")
	c.Contains([]int{1, 2}, 1, "expect to contain")
	c.Must.Contains("test", "es", "expect to contain")

	c.NotContains("hi", "es", "expect to contain")
	c.NotContains([]int{1, 2}, 3, "expect to contain")
	c.Must.NotContains("hi", "es", "expect to contain")
}

func TestLen(t *testing.T) {
	c := New(t)

	c.Len([]int{1, 2}, 2, "expect length")
	c.Must.Len([]int{1, 2}, 2, "expect length")

	c.NotLen([]int{1, 2}, 1, "expect no length")
	c.Must.NotLen([]int{1, 2}, 1, "expect no length")
}

func TestIs(t *testing.T) {
	c := New(t)

	c.Is(1, 2, "expect same types")
	c.Must.Is(1, 2, "expect same types")

	c.NotIs(1, c, "expect different types")
	c.NotIs(1, 2.0, "expect different types")
	c.Must.NotIs(1, c, "expect different types")
}

func TestNil(t *testing.T) {
	c := New(t)

	c.NotNil(errors.New("test"), "expect not nil")
	c.Must.NotNil(errors.New("test"), "expect not nil")

	c.Nil(nil, "expect nil")
	c.Must.Nil(nil, "expect nil")
}

func TestPanic(t *testing.T) {
	c := New(t)

	panics := func() {
		panic("oh noez!")
	}

	c.Panics(panics, "expect panic")
	c.Must.Panics(panics, "expect panic")

	c.NotPanics(func() {}, "expect no panic")
	c.Must.NotPanics(func() {}, "expect no panic")
}

func TestUntil(t *testing.T) {
	c := New(t)

	i := 0
	c.Until(time.Second, func() bool { i++; return i > 10 }, "failed")
}

func TestUntilNil(t *testing.T) {
	c := New(t)

	c.UntilNil(100, func() error {
		return nil
	})

	i := 0
	c.UntilNil(100, func() error {
		i++

		if i < 50 {
			return errors.New("merp")
		}

		return nil
	})
}
