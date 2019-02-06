package check

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type testC struct {
	*C
	next chan bool
}

func newTest(t *testing.T) testC {
	c := New(t)

	next := make(chan bool, 1)
	c.assert.onFail = func(msgArgs ...interface{}) {
		next <- false
	}

	c.Must.onFail = func(msgArgs ...interface{}) {
		next <- false
	}

	return testC{
		C:    c,
		next: next,
	}
}

func (c testC) expectOK(res bool) {
	if !res {
		c.Helper()
		c.Error("expected test to return true")
	}

	select {
	case <-c.next:
		c.Helper()
		c.Error("expected previous test to be OK")

	default:
	}
}

func (c testC) expectFail(res bool) {
	if res {
		c.Helper()
		c.Error("expected test to return false")
	}

	select {
	case <-c.next:

	default:
		c.Helper()
		c.Error("expected previous test to fail")
	}
}

func TestBool(t *testing.T) {
	c := newTest(t)

	c.expectOK(c.True(true, "expect true"))
	c.expectOK(c.Truef(true, "expect %s", "true"))
	c.expectFail(c.True(false, "expect false"))
	c.expectFail(c.Truef(false, "expect %s", "false"))

	c.expectOK(c.False(false, "expect false"))
	c.expectOK(c.Falsef(false, "expect %s", "false"))
	c.expectFail(c.False(true, "expect false"))
	c.expectFail(c.Falsef(true, "expect %s", "false"))
}

func TestEqual(t *testing.T) {
	c := newTest(t)

	tests := []struct {
		e     interface{}
		g     interface{}
		equal bool
	}{
		{
			e:     1,
			g:     1,
			equal: true,
		},
		{
			e:     int(1),
			g:     int64(1),
			equal: false,
		},
		{
			e:     1.0,
			g:     int64(1),
			equal: false,
		},
		{
			e:     1,
			g:     2,
			equal: false,
		},
		{
			e:     "some long string",
			g:     "another long string",
			equal: false,
		},
		{
			e:     map[string]string{"a": "a", "b": "b"},
			g:     map[string]string{"a": "a", "b": "b"},
			equal: true,
		},
		{
			e:     map[string]string{"a": "a", "b": "b"},
			g:     map[string]string{"a": "a"},
			equal: false,
		},
		{
			e:     1.0,
			g:     2,
			equal: false,
		},
		{
			e:     nil,
			g:     []byte(nil),
			equal: false,
		},
		{
			e:     []byte(nil),
			g:     []byte(nil),
			equal: true,
		},
		{
			e:     []byte(nil),
			g:     nil,
			equal: false,
		},
		{
			e:     nil,
			g:     nil,
			equal: true,
		},
		{
			e:     (io.Reader)((*bytes.Buffer)(nil)),
			g:     nil,
			equal: false,
		},
		{
			e:     nil,
			g:     1,
			equal: false,
		},
	}

	for i, test := range tests {
		if test.equal {
			c.expectOK(c.Equalf(test.g, test.e, "%d", i))
			c.expectOK(c.Must.Equalf(test.g, test.e, "%d", i))
			c.expectOK(c.Equalf(test.e, test.g, "%d", i))
			c.expectOK(c.Must.Equalf(test.e, test.g, "%d", i))
			c.expectFail(c.NotEqualf(test.g, test.e, "%d", i))
			c.expectFail(c.Must.NotEqualf(test.g, test.e, "%d", i))
			c.expectFail(c.NotEqualf(test.e, test.g, "%d", i))
			c.expectFail(c.Must.NotEqualf(test.e, test.g, "%d", i))
		} else {
			c.expectOK(c.NotEqualf(test.g, test.e, "%d", i))
			c.expectOK(c.Must.NotEqualf(test.g, test.e, "%d", i))
			c.expectOK(c.NotEqualf(test.e, test.g, "%d", i))
			c.expectOK(c.Must.NotEqualf(test.e, test.g, "%d", i))
			c.expectFail(c.Equalf(test.g, test.e, "%d", i))
			c.expectFail(c.Must.Equalf(test.g, test.e, "%d", i))
			c.expectFail(c.Equalf(test.e, test.g, "%d", i))
			c.expectFail(c.Must.Equalf(test.e, test.g, "%d", i))
		}
	}
}

func TestEqualExtras(t *testing.T) {
	c := newTest(t)

	c.expectOK(c.Equal(1, 1))
	c.expectOK(c.NotEqual(1, 2))

	// Interface ints vs untyped ints don't compare nicely
	// v := uint64(1)
	// c.expectOK(c.Equal(1, v))
}

func TestContains(t *testing.T) {
	c := newTest(t)

	tests := []struct {
		iter     interface{}
		el       interface{}
		contains bool
	}{
		{
			iter:     "some string",
			el:       "me st",
			contains: true,
		},
		{
			iter:     map[string]int{"test": 1, "test2": 2},
			el:       nil,
			contains: false,
		},
		{
			iter:     map[string]int{"test": 1, "test2": 2},
			el:       "test",
			contains: true,
		},
		{
			iter:     map[string]int{"test": 1, "test2": 2},
			el:       "test123",
			contains: false,
		},
		{
			iter:     []int{1, 2},
			el:       1,
			contains: true,
		},
		{
			iter:     []int{1, 2},
			el:       3,
			contains: false,
		},
		{
			iter:     []string{"test", "test2"},
			el:       "test2",
			contains: true,
		},
		{
			iter:     []string{"test", "test2"},
			el:       "test3",
			contains: false,
		},
	}

	for i, test := range tests {
		if test.contains {
			c.expectOK(c.Containsf(test.iter, test.el, "%d", i))
			c.expectFail(c.NotContainsf(test.iter, test.el, "%d", i))
		} else {
			c.expectFail(c.Containsf(test.iter, test.el, "%d", i))
			c.expectOK(c.NotContainsf(test.iter, test.el, "%d", i))
		}
	}

	c.expectFail(c.Contains(nil, 1))
	c.expectFail(c.Contains(123, 1))
	c.expectFail(c.NotContains(nil, 1))
	c.expectFail(c.NotContains(123, 1))
}

func TestLen(t *testing.T) {
	c := newTest(t)

	tests := []struct {
		iter interface{}
		n    int
	}{
		{
			iter: "",
			n:    0,
		},
		{
			iter: "test",
			n:    4,
		},
		{
			iter: []int{},
			n:    0,
		},
		{
			iter: []int{1, 2},
			n:    2,
		},
		{
			iter: map[int]int{1: 1, 2: 2},
			n:    2,
		},
	}

	for i, test := range tests {
		c.expectOK(c.Lenf(test.iter, test.n, "%d", i))
		c.expectFail(c.NotLenf(test.iter, test.n, "%d", i))
	}

	c.expectFail(c.Len(nil, 1))
	c.expectFail(c.Len(123, 1))
	c.expectFail(c.NotLen(nil, 1))
	c.expectFail(c.NotLen(123, 1))
}

func TestNil(t *testing.T) {
	c := newTest(t)

	c.expectOK(c.Nilf(nil, "%s", "nil"))
	c.expectFail(c.Nilf(errors.New("test"), "not %s", "nil"))

	c.expectOK(c.NotNilf(errors.New("test"), "not %s", "nil"))
	c.expectFail(c.NotNilf(nil, "%s", "nil"))
}

func TestPanic(t *testing.T) {
	c := newTest(t)

	panics := func() { panic("oh noez!") }
	noop := func() {}

	c.expectOK(c.Panicsf(panics, "%s", "panics"))
	c.expectFail(c.Panicsf(noop, "not %s", "panics"))

	c.expectOK(c.NotPanicsf(noop, "not %s", "panics"))
	c.expectFail(c.NotPanicsf(panics, "%s", "panics"))
}

func TestUntil(t *testing.T) {
	c := newTest(t)

	i := 0
	c.expectOK(c.Untilf(1000, func() bool { i++; return i > 10 }, "%s", "ok"))
	c.expectFail(c.Untilf(1000, func() bool { return false }, "not %s", "ok"))
}

func TestUntilNil(t *testing.T) {
	c := newTest(t)

	c.expectOK(c.UntilNilf(100, func() error { return nil }, "%s", "ok"))

	i := 0
	c.expectOK(c.UntilNilf(100, func() error {
		i++

		if i < 50 {
			return errors.New("merp")
		}

		return nil
	}, "%s", "ok"))

	c.expectFail(c.UntilNilf(
		100,
		func() error { return errors.New("merp") },
		"not %s", "ok"))
}
