package check

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/thatguystone/cog/stringc"
)

type assert struct {
	testing.TB
	onFail func(msgArgs ...interface{})
}

func (a assert) errorf(format string, args ...interface{}) {
	a.Helper()
	a.onFail(fmt.Sprintf(format, args...))
}

// True checks that the given bool is true.
func (a assert) True(cond bool, msgArgs ...interface{}) bool {
	if !cond {
		a.Helper()
		a.errorf("%s\n"+
			"Bool check failed: expected true",
			fmt.Sprint(msgArgs...))
	}

	return cond
}

// Truef checks that the given bool is true.
func (a assert) Truef(cond bool, format string, args ...interface{}) bool {
	a.Helper()
	return a.True(cond, fmt.Sprintf(format, args...))
}

// False checks that the given bool is false.
func (a assert) False(cond bool, msgArgs ...interface{}) bool {
	if cond {
		a.Helper()
		a.errorf("%s\n"+
			"Bool check failed: expected false",
			fmt.Sprint(msgArgs...))
	}

	return !cond
}

// Falsef checks that the given bool is false.
func (a assert) Falsef(cond bool, format string, args ...interface{}) bool {
	a.Helper()
	return a.False(cond, fmt.Sprintf(format, args...))
}

func (a assert) equal(g, e interface{}) bool {
	return reflect.DeepEqual(g, e)
}

// Equal compares two things, ensuring that they are equal to each other. `e` is
// the expected value; `g` is the value you got somewhere else.
func (a assert) Equal(g, e interface{}, msgArgs ...interface{}) bool {
	if !a.equal(g, e) {
		diff := diff(g, e)

		if diff != "" {
			diff = "\n\nDiff:\n" + stringc.Indent(diff, spewConfig.Indent)
		}

		g, e := fmtVals(g, e)

		a.Helper()
		a.errorf("%s\n"+
			"Expected: `%+v`\n"+
			"       == `%+v`%s",
			fmt.Sprint(msgArgs...),
			g, e, diff)
		return false
	}

	return true
}

// Equal compares two things, ensuring that they are equal to each other. `e` is
// the expected value; `g` is the value you got somewhere else.
func (a assert) Equalf(g, e interface{}, format string, args ...interface{}) bool {
	a.Helper()
	return a.Equal(g, e, fmt.Sprintf(format, args...))
}

// NotEqual is the opposite of Equal.
func (a assert) NotEqual(g, e interface{}, msgArgs ...interface{}) bool {
	if a.equal(g, e) {
		a.Helper()
		a.errorf("%s\n"+
			"Expected: `%+v`\n"+
			"       != `%+v`",
			fmt.Sprint(msgArgs...),
			g, e)
		return false
	}

	return true
}

// NotEqualf is the opposite of Equalf.
func (a assert) NotEqualf(
	g, e interface{},
	format string,
	args ...interface{},
) bool {
	a.Helper()
	return a.NotEqual(g, e, fmt.Sprintf(format, args...))
}

func getLen(v interface{}) (n int, ok bool) {
	defer func() { recover() }()

	n = reflect.ValueOf(v).Len()
	ok = true
	return
}

// Len checks that the length of the given v is l.
func (a assert) Len(v interface{}, l int, msgArgs ...interface{}) (eq bool) {
	n, ok := getLen(v)
	if !ok {
		a.Helper()
		a.errorf("%s\n"+
			"%+v is not iterable, cannot check length",
			fmt.Sprint(msgArgs...),
			v)
		return false
	}

	a.Helper()
	return a.Equal(n, l, msgArgs...)
}

// Lenf checks that the length of the given v is l.
func (a assert) Lenf(
	v interface{},
	l int,
	format string,
	args ...interface{},
) (eq bool) {
	a.Helper()
	return a.Len(v, l, fmt.Sprintf(format, args...))
}

// NotLen is the opposite of Len.
func (a assert) NotLen(v interface{}, l int, msgArgs ...interface{}) (eq bool) {
	n, ok := getLen(v)
	if !ok {
		a.Helper()
		a.errorf("%s\n"+
			"%+v is not iterable, cannot check length",
			fmt.Sprint(msgArgs...),
			v)
		return false
	}

	a.Helper()
	return a.NotEqual(n, l, msgArgs...)
}

// NotLenf is the opposite of Lenf.
func (a assert) NotLenf(
	v interface{},
	l int,
	format string,
	args ...interface{},
) (eq bool) {
	a.Helper()
	return a.NotLen(v, l, fmt.Sprintf(format, args...))
}

func (a assert) contains(iter, el interface{}) (found, ok bool) {
	ok = true

	iv := reflect.ValueOf(iter)
	ev := reflect.ValueOf(el)
	defer func() {
		if err := recover(); err != nil {
			ok = false
		}
	}()

	if iv.Kind() == reflect.String {
		found = strings.Contains(iv.String(), ev.String())
		return
	}

	if iv.Kind() == reflect.Map {
		for _, k := range iv.MapKeys() {
			if a.equal(k.Interface(), el) {
				found = true
				return
			}
		}

		return
	}

	for i := 0; i < iv.Len(); i++ {
		if a.equal(iv.Index(i).Interface(), el) {
			found = true
			return
		}
	}

	return
}

// Contains checks that iter contains v.
//
// The following checks are done:
//    0) If iter is a string, falls back to strings.Contains.
//    1) If a map, checks if the map contains key v.
//    2) If iter is a slice/array, checks if any element in iter equals v.
func (a assert) Contains(iter, v interface{}, msgArgs ...interface{}) bool {
	found, ok := a.contains(iter, v)

	if !ok {
		a.Helper()
		a.errorf("%s\n"+
			"%+v is not iterable; contain check failed",
			fmt.Sprint(msgArgs...),
			v)
		return false
	}

	if !found {
		a.Helper()
		a.errorf("%s\n"+
			"%#v does not contain %#v",
			fmt.Sprint(msgArgs...),
			iter,
			v)
		return false
	}

	return true
}

// Containsf checks that iter contains v.
//
// The following checks are done:
//    0) If iter is a string, falls back to strings.Contains.
//    1) If a map, checks if the map contains key v.
//    2) If iter is a slice/array, checks if any element in iter equals v.
func (a assert) Containsf(
	iter, v interface{},
	format string,
	args ...interface{},
) bool {
	a.Helper()
	return a.Contains(iter, v, fmt.Sprintf(format, args...))
}

// NotContains is the opposite of Contains.
func (a assert) NotContains(iter, v interface{}, msgArgs ...interface{}) bool {
	found, ok := a.contains(iter, v)

	if !ok {
		a.Helper()
		a.errorf("%s\n"+
			"%+v is not iterable; contain check failed",
			fmt.Sprint(msgArgs...),
			v)
		return false
	}

	if found {
		a.Helper()
		a.errorf("%s\n"+
			"%#v contains %#v",
			fmt.Sprint(msgArgs...),
			iter,
			v)
		return false
	}

	return true
}

// NotContainsf is the opposite of Containsf.
func (a assert) NotContainsf(
	iter, v interface{},
	format string,
	args ...interface{},
) bool {
	a.Helper()
	return a.NotContains(iter, v, fmt.Sprintf(format, args...))
}

// Nil ensures that g is nil. This is a strict equality check.
func (a assert) Nil(g interface{}, msgArgs ...interface{}) bool {
	if g != nil {
		a.Helper()
		a.errorf("%s\n"+
			"Expected nil, got: `%+v`",
			fmt.Sprint(msgArgs...),
			g)
		return false
	}

	return true
}

// Nilf ensures that g is nil. This is a strict equality check.
func (a assert) Nilf(g interface{}, format string, args ...interface{}) bool {
	a.Helper()
	return a.Nil(g, fmt.Sprintf(format, args...))
}

// NotNil is the opposite of Nil.
func (a assert) NotNil(g interface{}, msgArgs ...interface{}) bool {
	if g == nil {
		a.Helper()
		a.errorf("%s\n"+
			"Expected something, got nil",
			fmt.Sprint(msgArgs...))
		return false
	}

	return true
}

// NotNilf is the opposite of Nilf.
func (a assert) NotNilf(g interface{}, format string, args ...interface{}) bool {
	a.Helper()
	return a.NotNil(g, fmt.Sprintf(format, args...))
}

// Panics ensures that the given function panics.
func (a assert) Panics(fn func(), msgArgs ...interface{}) (ok bool) {
	defer func() { recover() }() // Can't just defer recover(), apparently
	ok = true
	fn()

	ok = false
	a.Helper()
	a.errorf("%s\n"+
		"Expected fn to panic; it did not.",
		fmt.Sprint(msgArgs...))

	return
}

// Panicsf ensures that the given function panics.
func (a assert) Panicsf(fn func(), format string, args ...interface{}) bool {
	a.Helper()
	return a.Panics(fn, fmt.Sprintf(format, args...))
}

// NotPanics ensures that the given function does not panic.
func (a assert) NotPanics(fn func(), msgArgs ...interface{}) (ok bool) {
	var r interface{}
	var stack string

	// Need to do the check in a different function, otherwise a.Helper()
	// doesn't work (it reports that the callsite is in some .s file)
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
		a.Helper()
		a.errorf("%s\n"+
			"Expected fn not to panic; got: %+v\n%s",
			fmt.Sprint(msgArgs...),
			r,
			stringc.Indent(stack, "\t"))
	}

	return ok
}

// NotPanicsf ensures that the given function does not panic.
func (a assert) NotPanicsf(fn func(), format string, args ...interface{}) bool {
	a.Helper()
	return a.NotPanics(fn, fmt.Sprintf(format, args...))
}

// Until polls the given function, a max of `iters` times, until it returns
// true.
func (a assert) Until(
	iters int,
	fn func() bool,
	msgArgs ...interface{},
) bool {
	for i := 0; i < iters; i++ {
		if fn() {
			return true
		}
	}

	a.errorf("%s\n"+
		"Polling for condition failed",
		fmt.Sprint(msgArgs...))
	return false
}

// Untilf polls the given function, a max of `iters` times, until it returns
// true.
func (a assert) Untilf(
	iters int,
	fn func() bool,
	format string,
	args ...interface{},
) bool {
	a.Helper()
	return a.Until(iters, fn, fmt.Sprintf(format, args...))
}

// UntilNil polls the given function, a max of `iters` times, until it doesn't
// return an error. This is mainly a helper used to exhaust error pathways when
// using an Errorer.
func (a assert) UntilNil(
	iters int,
	fn func() error,
	msgArgs ...interface{},
) bool {
	var err error

	for i := 0; i < iters; i++ {
		err = fn()
		if err == nil {
			return true
		}
	}

	a.Helper()
	a.errorf("%s\n"+
		"Func didn't succeed after %d tries, last err: %v",
		fmt.Sprint(msgArgs...),
		iters, err)
	return false
}

// UntilNilf polls the given function, a max of `iters` times, until it doesn't
// return an error. This is mainly a helper used to exhaust error pathways when
// using an Errorer.
func (a assert) UntilNilf(
	iters int,
	fn func() error,
	format string,
	args ...interface{},
) bool {
	a.Helper()
	return a.UntilNil(iters, fn, fmt.Sprintf(format, args...))
}
