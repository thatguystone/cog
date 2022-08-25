package check

import (
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/thatguystone/cog/textwrap"
)

type assert struct {
	helper func()
	fail   func(msgArgs ...any)
}

func newAssert(helper func(), fail func(msgArgs ...any)) assert {
	return assert{
		helper: helper,
		fail:   fail,
	}
}

func (a assert) error(userMessage, failReason string) {
	a.helper()
	a.fail(userMessage + "\n" + failReason)
}

// True checks that the given bool is true.
func (a assert) True(cond bool) bool {
	a.helper()
	return a.Truef(cond, "")
}

// Truef checks that the given bool is true.
func (a assert) Truef(cond bool, format string, args ...any) bool {
	if !cond {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			"Bool check failed: expected true",
		)
	}

	return cond
}

// False checks that the given bool is false.
func (a assert) False(cond bool) bool {
	a.helper()
	return a.Falsef(cond, "")
}

// Falsef checks that the given bool is false.
func (a assert) Falsef(cond bool, format string, args ...any) bool {
	if cond {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			"Bool check failed: expected false",
		)
	}

	return !cond
}

func (a assert) equal(g, e any) bool {
	return reflect.DeepEqual(g, e)
}

// Equal compares two things, ensuring that they are equal to each other. `e` is
// the expected value; `g` is the value you got somewhere else.
func (a assert) Equal(g, e any) bool {
	a.helper()
	return a.Equalf(g, e, "")
}

// Equal compares two things, ensuring that they are equal to each other. `e` is
// the expected value; `g` is the value you got somewhere else.
func (a assert) Equalf(g, e any, format string, args ...any) bool {
	if !a.equal(g, e) {
		diff := diff(g, e)

		if diff != "" {
			diff = "\n\nDiff:\n" + textwrap.Indent(diff, spewConfig.Indent)
		}

		gf, ef := fmtVals(g, e)

		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf(""+
				"Expected: %s\n"+
				"       == %s%s",
				gf,
				ef,
				diff,
			),
		)
		return false
	}

	return true
}

// NotEqual is the opposite of Equal.
func (a assert) NotEqual(g, e any) bool {
	a.helper()
	return a.NotEqualf(g, e, "")
}

// NotEqualf is the opposite of Equalf.
func (a assert) NotEqualf(g, e any, format string, args ...any) bool {
	if a.equal(g, e) {
		gf, ef := fmtVals(g, e)

		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf(""+
				"Expected: %s\n"+
				"       != %s",
				gf,
				ef,
			),
		)
		return false
	}

	return true
}

func (a assert) hasKey(m, k any) (found, ok bool) {
	defer func() { recover() }()

	mi := reflect.ValueOf(m).MapRange()

	for mi.Next() {
		if a.equal(mi.Key().Interface(), k) {
			found = true
			break
		}
	}

	ok = true
	return
}

func (a assert) hasKeyf(
	shouldContain bool,
	m any,
	k any,
	format string,
	args ...any,
) bool {
	found, ok := a.hasKey(m, k)
	if !ok {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%+v is not a map, cannot check for key", k),
		)
		return false
	}

	if shouldContain && !found {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v does not contain key %#v", m, k),
		)
		return false
	}

	if !shouldContain && found {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v unexpectedly contains key %#v", m, k),
		)
		return false
	}

	return true
}

// hasKey checks that map m contains k key
func (a assert) HasKey(m, k any) bool {
	a.helper()
	return a.HasKeyf(m, k, "")
}

// HasKeyf checks that map m contains k key
func (a assert) HasKeyf(m, k any, format string, args ...any) bool {
	a.helper()
	return a.hasKeyf(true, m, k, format, args...)
}

// NotHasKey checks that map m does not contain k key
func (a assert) NotHasKey(m, k any) bool {
	a.helper()
	return a.NotHasKeyf(m, k, "")
}

// NotHasKeyf checks that map m does not contain k key
func (a assert) NotHasKeyf(m, k any, format string, args ...any) bool {
	a.helper()
	return a.hasKeyf(false, m, k, format, args...)
}

func (a assert) hasVal(iter, el any) (found, ok bool) {
	defer func() { recover() }()

	iv := reflect.ValueOf(iter)

	switch iv.Kind() {
	case reflect.Map:
		mi := iv.MapRange()
		for mi.Next() {
			if a.equal(mi.Value().Interface(), el) {
				found = true
				break
			}
		}

	case reflect.Array, reflect.Slice:
		for i := 0; i < iv.Len(); i++ {
			if a.equal(iv.Index(i).Interface(), el) {
				found = true
				break
			}
		}

	default:
		return
	}

	ok = true
	return
}

func (a assert) hasValf(
	shouldContain bool,
	iter any,
	v any,
	format string,
	args ...any,
) bool {
	found, ok := a.hasVal(iter, v)
	if !ok {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%+v is not iterable, cannot check for value", iter),
		)
		return false
	}

	if shouldContain && !found {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v does not contain value %#v", iter, v),
		)
		return false
	}

	if !shouldContain && found {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v unexpectedly contains value %#v", iter, v),
		)
		return false
	}

	return true
}

// HasVal checks that iter contains value v.
//
// Iter must be one of: map, slice, or array
func (a assert) HasVal(iter, v any) bool {
	a.helper()
	return a.HasValf(iter, v, "")
}

// HasValf checks that iter contains value v.
//
// Iter must be one of: map, slice, or array
func (a assert) HasValf(iter, v any, format string, args ...any) bool {
	a.helper()
	return a.hasValf(true, iter, v, format, args...)
}

// NotHasVal is the opposite of HasVal.
func (a assert) NotHasVal(iter, v any) bool {
	a.helper()
	return a.NotHasValf(iter, v, "")
}

// NotHasValf is the opposite of HasValf.
func (a assert) NotHasValf(iter, v any, format string, args ...any) bool {
	a.helper()
	return a.hasValf(false, iter, v, format, args...)
}

// Nil ensures that g is nil. This is a strict equality check.
func (a assert) Nil(g any) bool {
	a.helper()
	return a.Nilf(g, "")
}

// Nilf ensures that g is nil. This is a strict equality check.
func (a assert) Nilf(g any, format string, args ...any) bool {
	if g != nil {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("Expected nil, got: `%+v`", g),
		)
		return false
	}

	return true
}

// NotNil is the opposite of Nil.
func (a assert) NotNil(g any) bool {
	a.helper()
	return a.NotNilf(g, "")
}

// NotNilf is the opposite of Nilf.
func (a assert) NotNilf(g any, format string, args ...any) bool {
	if g == nil {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			"Expected something, got nil",
		)
		return false
	}

	return true
}

// Panics ensures that the given function panics.
func (a assert) Panics(fn func()) bool {
	a.helper()
	return a.Panicsf(fn, "")
}

// Panicsf ensures that the given function panics.
func (a assert) Panicsf(fn func(), format string, args ...any) (ok bool) {
	defer func() { recover() }()
	ok = true
	fn()

	ok = false
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		"Expected fn to panic; it did not.",
	)

	return
}

// NotPanics ensures that the given function does not panic.
func (a assert) NotPanics(fn func()) bool {
	a.helper()
	return a.NotPanicsf(fn, "")
}

// NotPanicsf ensures that the given function does not panic.
func (a assert) NotPanicsf(fn func(), format string, args ...any) (ok bool) {
	var (
		r     any
		stack string
	)

	func() {
		defer func() {
			if !ok {
				r = recover()
				stack = string(debug.Stack())
			}
		}()

		fn()
		ok = true
	}()

	if !ok {
		a.helper()
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf(
				"Expected fn not to panic; got: %+v\n%s",
				r,
				textwrap.Indent(stack, "\t"),
			),
		)
	}

	return
}

// Until polls the given function, a max of `iters` times, until it returns
// true.
func (a assert) Until(iters int, fn func(i int) bool) bool {
	a.helper()
	return a.Untilf(iters, fn, "")
}

// Untilf polls the given function, a max of `iters` times, until it returns
// true.
func (a assert) Untilf(
	iters int,
	fn func(i int) bool,
	format string,
	args ...any,
) bool {
	for i := 0; i < iters; i++ {
		if fn(i) {
			return true
		}
	}

	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		"Polling for condition failed",
	)

	return false
}

// UntilNil polls the given function, a max of `iters` times, until it doesn't
// return an error. This is mainly a helper used to exhaust error pathways when
// using an Errorer.
func (a assert) UntilNil(iters int, fn func(i int) error) bool {
	a.helper()
	return a.UntilNilf(iters, fn, "")
}

// UntilNilf polls the given function, a max of `iters` times, until it doesn't
// return an error. This is mainly a helper used to exhaust error pathways when
// using an Errorer.
func (a assert) UntilNilf(
	iters int,
	fn func(i int) error,
	format string,
	args ...any,
) bool {
	var err error

	for i := 0; i < iters; i++ {
		err = fn(i)
		if err == nil {
			return true
		}
	}

	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		fmt.Sprintf(
			"Func didn't succeed after %d tries, last err: %v",
			iters,
			err,
		),
	)

	return false
}
