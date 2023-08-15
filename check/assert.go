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

func (a assert) failTruef(format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		"Bool check failed: expected true",
	)
}

// True checks that the given bool is true.
func (a assert) True(cond bool) (ok bool) {
	if !cond {
		a.helper()
		a.failTruef("")
		return false
	}

	return true
}

// Truef checks that the given bool is true.
func (a assert) Truef(cond bool, format string, args ...any) bool {
	if !cond {
		a.helper()
		a.failTruef(format, args...)
		return false
	}

	return true
}

func (a assert) failFalsef(format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		"Bool check failed: expected false",
	)
}

// False checks that the given bool is false.
func (a assert) False(cond bool) bool {
	if cond {
		a.helper()
		a.failFalsef("")
		return false
	}

	return true
}

// Falsef checks that the given bool is false.
func (a assert) Falsef(cond bool, format string, args ...any) bool {
	if cond {
		a.helper()
		a.failFalsef(format, args...)
		return false
	}

	return true
}

func (a assert) equal(g, e any) bool {
	return reflect.DeepEqual(g, e)
}

func (a assert) failEqualf(g, e any, format string, args ...any) {
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
}

// Equal compares two things, ensuring that they are equal to each other. e is
// the expected value; g is the value you got somewhere else.
func (a assert) Equal(g, e any) bool {
	if !a.equal(g, e) {
		a.helper()
		a.failEqualf(g, e, "")
		return false
	}

	return true
}

// Equal compares two things, ensuring that they are equal to each other. e is
// the expected value; g is the value you got somewhere else.
func (a assert) Equalf(g, e any, format string, args ...any) bool {
	if !a.equal(g, e) {
		a.helper()
		a.failEqualf(g, e, format, args...)
		return false
	}

	return true
}

func (a assert) failNotEqualf(g, e any, format string, args ...any) {
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
}

// NotEqual is the opposite of Equal.
func (a assert) NotEqual(g, e any) bool {
	if a.equal(g, e) {
		a.helper()
		a.failNotEqualf(g, e, "")
		return false
	}

	return true
}

// NotEqualf is the opposite of Equalf.
func (a assert) NotEqualf(g, e any, format string, args ...any) bool {
	if a.equal(g, e) {
		a.helper()
		a.failNotEqualf(g, e, format, args...)
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

func (a assert) failHasKeyf(
	ok bool,
	m any,
	k any,
	format string,
	args ...any,
) {
	a.helper()

	if !ok {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%+v is not a map, cannot check for key", k),
		)
	} else {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v does not contain key %#v", m, k),
		)
	}
}

// hasKey checks that map m contains k key
func (a assert) HasKey(m, k any) bool {
	found, ok := a.hasKey(m, k)
	if !found || !ok {
		a.helper()
		a.failHasKeyf(ok, m, k, "")
		return false
	}

	return true
}

// HasKeyf checks that map m contains k key
func (a assert) HasKeyf(m, k any, format string, args ...any) bool {
	found, ok := a.hasKey(m, k)
	if !found || !ok {
		a.helper()
		a.failHasKeyf(ok, m, k, format, args...)
		return false
	}

	return true
}

func (a assert) failNotHasKeyf(
	ok bool,
	m any,
	k any,
	format string,
	args ...any,
) {
	a.helper()

	if !ok {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%+v is not a map, cannot check for key", k),
		)
	} else {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v unexpectedly contains key %#v", m, k),
		)
	}
}

// NotHasKey checks that map m does not contain k key
func (a assert) NotHasKey(m, k any) bool {
	found, ok := a.hasKey(m, k)
	if found || !ok {
		a.helper()
		a.failNotHasKeyf(ok, m, k, "")
		return false
	}

	return true
}

// NotHasKeyf checks that map m does not contain k key
func (a assert) NotHasKeyf(m, k any, format string, args ...any) bool {
	found, ok := a.hasKey(m, k)
	if found || !ok {
		a.helper()
		a.failNotHasKeyf(ok, m, k, format, args...)
		return false
	}

	return true
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

func (a assert) failHasValf(
	ok bool,
	iter any,
	v any,
	format string,
	args ...any,
) {
	a.helper()

	if !ok {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%+v is not iterable, cannot check for value", iter),
		)
	} else {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v does not contain value %#v", iter, v),
		)
	}
}

// HasVal checks that iter contains value v.
//
// Iter must be one of: map, slice, or array
func (a assert) HasVal(iter, v any) bool {
	found, ok := a.hasVal(iter, v)
	if !found || !ok {
		a.helper()
		a.failHasValf(ok, iter, v, "")
		return false
	}

	return true
}

// HasValf checks that iter contains value v.
//
// Iter must be one of: map, slice, or array
func (a assert) HasValf(iter, v any, format string, args ...any) bool {
	found, ok := a.hasVal(iter, v)
	if !found || !ok {
		a.helper()
		a.failHasValf(ok, iter, v, format, args...)
		return false
	}

	return true
}

func (a assert) failNotHasValf(
	ok bool,
	iter any,
	v any,
	format string,
	args ...any,
) {
	a.helper()

	if !ok {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%+v is not iterable, cannot check for value", iter),
		)
	} else {
		a.error(
			fmt.Sprintf(format, args...),
			fmt.Sprintf("%#v unexpectedly contains value %#v", iter, v),
		)
	}
}

// NotHasVal is the opposite of HasVal.
func (a assert) NotHasVal(iter, v any) bool {
	found, ok := a.hasVal(iter, v)
	if found || !ok {
		a.helper()
		a.failNotHasValf(ok, iter, v, "")
		return false
	}

	return true
}

// NotHasValf is the opposite of HasValf.
func (a assert) NotHasValf(iter, v any, format string, args ...any) bool {
	found, ok := a.hasVal(iter, v)
	if found || !ok {
		a.helper()
		a.failNotHasValf(ok, iter, v, format, args...)
		return false
	}

	return true
}

func (a assert) failNilf(g any, format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		fmt.Sprintf("Expected nil, got: `%+v`", g),
	)
}

// Nil ensures that g is nil. This is a strict equality check.
func (a assert) Nil(g any) bool {
	if g != nil {
		a.helper()
		a.failNilf(g, "")
		return false
	}

	return true
}

// Nilf ensures that g is nil. This is a strict equality check.
func (a assert) Nilf(g any, format string, args ...any) bool {
	if g != nil {
		a.helper()
		a.failNilf(g, format, args...)
		return false
	}

	return true
}

func (a assert) failNotNilf(format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		"Expected something, got nil",
	)
}

// NotNil is the opposite of Nil.
func (a assert) NotNil(g any) bool {
	if g == nil {
		a.helper()
		a.failNotNilf("")
		return false
	}

	return true
}

// NotNilf is the opposite of Nilf.
func (a assert) NotNilf(g any, format string, args ...any) bool {
	if g == nil {
		a.helper()
		a.failNotNilf(format, args...)
		return false
	}

	return true
}

type paniced struct {
	recovered any
	stack     string
	didPanic  bool
}

func checkPanic(fn func()) (p paniced) {
	defer func() {
		p.recovered = recover()

		if p.didPanic {
			p.stack = string(debug.Stack())
		}
	}()

	p.didPanic = true
	fn()
	p.didPanic = false

	return
}

func (a assert) failPanicsf(format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		"Expected fn to panic; it did not.",
	)
}

// Panics ensures that the given function panics.
func (a assert) Panics(fn func()) bool {
	if !checkPanic(fn).didPanic {
		a.helper()
		a.failPanicsf("")
		return false
	}

	return true
}

// Panicsf ensures that the given function panics.
func (a assert) Panicsf(fn func(), format string, args ...any) (ok bool) {
	if !checkPanic(fn).didPanic {
		a.helper()
		a.failPanicsf(format, args...)
		return false
	}

	return true
}

// PanicsWith ensures that the given function panics with the given value.
func (a assert) PanicsWith(recovers any, fn func()) bool {
	p := checkPanic(fn)
	if !p.didPanic {
		a.helper()
		a.failPanicsf("")
		return false
	}

	if !a.equal(p.recovered, recovers) {
		a.helper()
		a.failEqualf(p.recovered, recovers, "")
		a.fail(p.stack)
		return false
	}

	return true
}

// PanicsWithf ensures that the given function panics with the given value.
func (a assert) PanicsWithf(recovers any, fn func(), format string, args ...any) bool {
	p := checkPanic(fn)
	if !p.didPanic {
		a.helper()
		a.failPanicsf(format, args...)
		return false
	}

	if !a.equal(p.recovered, recovers) {
		a.helper()
		a.failEqualf(p.recovered, recovers, format, args...)
		a.fail(p.stack)
		return false
	}

	return true
}

func (a assert) failNotPanicsf(p paniced, format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		fmt.Sprintf(
			"Expected fn not to panic; got: %+v\n%s",
			p.recovered,
			textwrap.Indent(p.stack, "\t"),
		),
	)
}

// NotPanics ensures that the given function does not panic.
func (a assert) NotPanics(fn func()) bool {
	p := checkPanic(fn)
	if p.didPanic {
		a.helper()
		a.failNotPanicsf(p, "")
		return false
	}

	return true
}

// NotPanicsf ensures that the given function does not panic.
func (a assert) NotPanicsf(fn func(), format string, args ...any) (ok bool) {
	p := checkPanic(fn)
	if p.didPanic {
		a.helper()
		a.failNotPanicsf(p, format, args...)
		return false
	}

	return true
}

func (a assert) until(iters int, fn func(i int) bool) bool {
	for i := 0; i < iters; i++ {
		if fn(i) {
			return true
		}
	}

	return false
}

func (a assert) failUntilf(iters int, format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		fmt.Sprintf(
			"Polling for condition failed, gave up after %d tries",
			iters,
		),
	)
}

// Until polls the given function, a max of iters times, until it returns
// true.
func (a assert) Until(iters int, fn func(i int) bool) bool {
	if !a.until(iters, fn) {
		a.helper()
		a.failUntilf(iters, "")
		return false
	}

	return true
}

// Untilf polls the given function, a max of iters times, until it returns
// true.
func (a assert) Untilf(
	iters int,
	fn func(i int) bool,
	format string,
	args ...any,
) bool {
	if !a.until(iters, fn) {
		a.helper()
		a.failUntilf(iters, format, args...)
		return false
	}

	return true
}

func (a assert) untilNil(iters int, fn func(i int) error) (err error) {
	for i := 0; i < iters; i++ {
		err = fn(i)
		if err == nil {
			return
		}
	}

	return
}

func (a assert) failUntilNilf(iters int, err error, format string, args ...any) {
	a.helper()
	a.error(
		fmt.Sprintf(format, args...),
		fmt.Sprintf(
			"Func didn't succeed after %d tries, last err: %v",
			iters,
			err,
		),
	)
}

// UntilNil polls the given function, a max of iters times, until it doesn't
// return an error. This is mainly a helper used to exhaust error pathways when
// using an Errorer.
func (a assert) UntilNil(iters int, fn func(i int) error) bool {
	err := a.untilNil(iters, fn)
	if err != nil {
		a.helper()
		a.failUntilNilf(iters, err, "")
		return false
	}

	return true
}

// UntilNilf polls the given function, a max of iters times, until it doesn't
// return an error. This is mainly a helper used to exhaust error pathways when
// using an Errorer.
func (a assert) UntilNilf(
	iters int,
	fn func(i int) error,
	format string,
	args ...any,
) bool {
	err := a.untilNil(iters, fn)
	if err != nil {
		a.helper()
		a.failUntilNilf(iters, err, format, args...)
		return false
	}

	return true
}
