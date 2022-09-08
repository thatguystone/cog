package check

import (
	"errors"
	"fmt"
	"testing"
)

type testAssert struct {
	assert
	t      *testing.T
	called bool
	ok     bool
	msg    string
}

func newTestAssert(t *testing.T) *testAssert {
	ta := new(testAssert)
	ta.t = t
	ta.assert = newAssert(ta.helper, ta.fail)
	return ta
}

func (ta *testAssert) helper() {}

func (ta *testAssert) fail(msgArgs ...any) {
	if len(msgArgs) != 1 {
		panic(fmt.Errorf("unexpected msgArgs len: %d", len(msgArgs)))
	}

	ta.called = true
	ta.ok = false
	ta.msg = msgArgs[0].(string)
}

func (ta *testAssert) check(callRet, expect bool) {
	ta.t.Helper()

	if !expect && !ta.called {
		ta.t.Error("assert not called")
	}

	if callRet != expect {
		ta.t.Errorf("call return mismatch: %t != %t", callRet, expect)
	}

	ta.called = false
	ta.ok = false
	ta.msg = ""
}

func TestAssertTrue(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.True(true),
		true,
	)
	ta.check(
		ta.Truef(true, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.True(false),
		false,
	)
	ta.check(
		ta.Truef(false, "fmt %d", 1),
		false,
	)
}

func TestAssertFalse(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.False(false),
		true,
	)
	ta.check(
		ta.Falsef(false, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.False(true),
		false,
	)
	ta.check(
		ta.Falsef(true, "fmt %d", 1),
		false,
	)
}

func TestAssertEqual(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.Equal(1, 1),
		true,
	)
	ta.check(
		ta.Equalf(1, 1, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.Equal(1, 2),
		false,
	)
	ta.check(
		ta.Equalf(1, 2, "fmt %d", 1),
		false,
	)

	type m struct {
		a int
	}

	ta.check(
		ta.Equal(m{a: 1}, m{a: 1}),
		true,
	)
	ta.check(
		ta.Equalf(m{a: 1}, m{a: 1}, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.Equal(m{a: 1}, m{a: 2}),
		false,
	)
	ta.check(
		ta.Equalf(m{a: 1}, m{a: 2}, "fmt %d", 1),
		false,
	)
}

func TestAssertNotEqual(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.NotEqual(1, 2),
		true,
	)
	ta.check(
		ta.NotEqualf(1, 2, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotEqual(1, 1),
		false,
	)
	ta.check(
		ta.NotEqualf(1, 1, "fmt %d", 1),
		false,
	)

	type m struct {
		a int
	}

	ta.check(
		ta.NotEqual(m{a: 1}, m{a: 2}),
		true,
	)
	ta.check(
		ta.NotEqualf(m{a: 1}, m{a: 2}, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotEqual(m{a: 1}, m{a: 1}),
		false,
	)
	ta.check(
		ta.NotEqualf(m{a: 1}, m{a: 1}, "fmt %d", 1),
		false,
	)
}

func TestAssertHasKey(t *testing.T) {
	ta := newTestAssert(t)

	m := map[string]int{
		"a": 1,
		"b": 2,
	}

	ta.check(
		ta.HasKey(m, "a"),
		true,
	)
	ta.check(
		ta.HasKeyf(m, "a", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.HasKey(m, "b"),
		true,
	)
	ta.check(
		ta.HasKeyf(m, "b", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.HasKey(m, "1"),
		false,
	)
	ta.check(
		ta.HasKeyf(m, "1", "fmt %d", 1),
		false,
	)

	ta.check(
		ta.HasKey(m, 1),
		false,
	)
	ta.check(
		ta.HasKeyf(m, 1, "fmt %d", 1),
		false,
	)

	ta.check(
		ta.HasKey(1, 1),
		false,
	)
	ta.check(
		ta.HasKeyf(1, 1, "fmt %d", 1),
		false,
	)
}

func TestAssertNotHasKey(t *testing.T) {
	ta := newTestAssert(t)

	m := map[string]int{
		"a": 1,
		"b": 2,
	}

	ta.check(
		ta.NotHasKey(m, "1"),
		true,
	)
	ta.check(
		ta.NotHasKeyf(m, "1", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotHasKey(m, 1),
		true,
	)
	ta.check(
		ta.NotHasKeyf(m, 1, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotHasKey(m, "a"),
		false,
	)
	ta.check(
		ta.NotHasKeyf(m, "a", "fmt %d", 1),
		false,
	)

	ta.check(
		ta.NotHasKey(m, "b"),
		false,
	)
	ta.check(
		ta.NotHasKeyf(m, "b", "fmt %d", 1),
		false,
	)

	ta.check(
		ta.NotHasKey(1, 1),
		false,
	)
	ta.check(
		ta.NotHasKeyf(1, 1, "fmt %d", 1),
		false,
	)
}

func TestAssertHasVal(t *testing.T) {
	ta := newTestAssert(t)

	m := map[string]int{
		"a": 1,
		"b": 2,
	}

	ta.check(
		ta.HasVal(m, 1),
		true,
	)
	ta.check(
		ta.HasValf(m, 1, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.HasVal(m, 2),
		true,
	)
	ta.check(
		ta.HasValf(m, 2, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.HasVal(m, "1"),
		false,
	)
	ta.check(
		ta.HasValf(m, "1", "fmt %d", 1),
		false,
	)

	ta.check(
		ta.HasVal(m, "a"),
		false,
	)
	ta.check(
		ta.HasValf(m, "a", "fmt %d", 1),
		false,
	)

	s := []int{1, 2}

	ta.check(
		ta.HasVal(s, 1),
		true,
	)
	ta.check(
		ta.HasValf(s, 1, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.HasVal(s, 2),
		true,
	)
	ta.check(
		ta.HasValf(s, 2, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.HasVal(s, "1"),
		false,
	)
	ta.check(
		ta.HasValf(s, "1", "fmt %d", 1),
		false,
	)

	ta.check(
		ta.HasVal(s, "a"),
		false,
	)
	ta.check(
		ta.HasValf(s, "a", "fmt %d", 1),
		false,
	)

	ta.check(
		ta.HasVal(1, 1),
		false,
	)
	ta.check(
		ta.HasValf(1, 1, "fmt %d", 1),
		false,
	)
}

func TestAssertNotHasVal(t *testing.T) {
	ta := newTestAssert(t)

	m := map[string]int{
		"a": 1,
		"b": 2,
	}

	ta.check(
		ta.NotHasVal(m, "1"),
		true,
	)
	ta.check(
		ta.NotHasValf(m, "1", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotHasVal(m, "a"),
		true,
	)
	ta.check(
		ta.NotHasValf(m, "a", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotHasVal(m, 1),
		false,
	)
	ta.check(
		ta.NotHasValf(m, 1, "fmt %d", 1),
		false,
	)

	ta.check(
		ta.NotHasVal(m, 2),
		false,
	)
	ta.check(
		ta.NotHasValf(m, 2, "fmt %d", 1),
		false,
	)

	s := []int{1, 2}

	ta.check(
		ta.NotHasVal(s, "1"),
		true,
	)
	ta.check(
		ta.NotHasValf(s, "1", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotHasVal(s, "a"),
		true,
	)
	ta.check(
		ta.NotHasValf(s, "a", "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotHasVal(s, 1),
		false,
	)
	ta.check(
		ta.NotHasValf(s, 1, "fmt %d", 1),
		false,
	)

	ta.check(
		ta.NotHasVal(s, 2),
		false,
	)
	ta.check(
		ta.NotHasValf(s, 2, "fmt %d", 1),
		false,
	)

	ta.check(
		ta.NotHasVal(1, 1),
		false,
	)
	ta.check(
		ta.NotHasValf(1, 1, "fmt %d", 1),
		false,
	)
}

func TestAssertNil(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.Nil(nil),
		true,
	)
	ta.check(
		ta.Nilf(nil, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.Nil(1),
		false,
	)
	ta.check(
		ta.Nilf(1, "fmt %d", 1),
		false,
	)
}

func TestAssertNotNil(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.NotNil(1),
		true,
	)
	ta.check(
		ta.NotNilf(1, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotNil(nil),
		false,
	)
	ta.check(
		ta.NotNilf(nil, "fmt %d", 1),
		false,
	)
}

func TestAssertPanics(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.Panics(func() { panic("panic") }),
		true,
	)
	ta.check(
		ta.Panicsf(func() { panic("panic") }, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.Panics(func() {}),
		false,
	)
	ta.check(
		ta.Panicsf(func() {}, "fmt %d", 1),
		false,
	)
}

func TestAssertNotPanics(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.NotPanics(func() {}),
		true,
	)
	ta.check(
		ta.NotPanicsf(func() {}, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.NotPanics(func() { panic("panic") }),
		false,
	)
	ta.check(
		ta.NotPanicsf(func() { panic("panic") }, "fmt %d", 1),
		false,
	)
}

func TestAssertUntil(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.Until(100, func(i int) bool { return i == 50 }),
		true,
	)
	ta.check(
		ta.Untilf(100, func(i int) bool { return i == 50 }, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.Until(100, func(i int) bool { return false }),
		false,
	)
	ta.check(
		ta.Untilf(100, func(i int) bool { return false }, "fmt %d", 1),
		false,
	)
}

func TestAssertUntilNil(t *testing.T) {
	ta := newTestAssert(t)

	ta.check(
		ta.UntilNil(100, func(i int) error {
			if i == 50 {
				return nil
			}

			return errors.New("error")
		}),
		true,
	)
	ta.check(
		ta.UntilNilf(100, func(i int) error {
			if i == 50 {
				return nil
			}

			return errors.New("error")
		}, "fmt %d", 1),
		true,
	)

	ta.check(
		ta.UntilNil(100, func(i int) error {
			return errors.New("error")
		}),
		false,
	)
	ta.check(
		ta.UntilNilf(100, func(i int) error {
			return errors.New("error")
		}, "fmt %d", 1),
		false,
	)
}
