package check

import (
	"fmt"
	"math"
	"math/big"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/iheartradio/cog/stack"
)

// Can't really get 100% coverage on this file: doing so causes the tests to
// fail.
//gocovr:skip-file

func format(msg ...interface{}) string {
	if msg == nil || len(msg) == 0 {
		return ""
	} else if len(msg) == 1 {
		return msg[0].(string)
	} else {
		return fmt.Sprintf(msg[0].(string), msg[1:]...)
	}
}

func callerInfo() string {
	var in string

	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			return "???:1"
		}

		file = path.Base(file)
		if file == "assert.go" {
			continue
		}

		fn := runtime.FuncForPC(pc)
		if strings.Contains(fn.Name(), ".Test") {
			if in == "" {
				return fmt.Sprintf("%s:%d", file, line)
			}

			return fmt.Sprintf("%s (from %s:%d)", in, file, line)
		}

		if in == "" {
			in = fmt.Sprintf("%s:%d", file, line)
		}
	}
}

func (c *C) getInt(e interface{}) (*big.Int, bool) {
	i := big.NewInt(0)

	v := reflect.ValueOf(e)
	switch v.Kind() {
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:

		i.SetInt64(v.Int())

	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:

		i.SetUint64(v.Uint())

	default:
		return nil, false
	}

	return i, true
}

func (c *C) intEqual(e, g interface{}) bool {
	ex, ok1 := c.getInt(e)
	gx, ok2 := c.getInt(g)
	if !ok1 || !ok2 {
		return false
	}

	return ex.Cmp(gx) == 0
}

func (c *C) floatingEqual(e, g interface{}) bool {
	fe, ok := e.(float64)
	if !ok {
		return false
	}

	fg, ok := g.(float64)
	if !ok {
		return false
	}

	min := math.Min(fe, fg)
	max := math.Max(fe, fg)

	return math.Nextafter(min, max) == max
}

func (c *C) equal(e, g interface{}) bool {
	if e == nil {
		if e == g {
			return true
		}

		v := reflect.ValueOf(g)
		k := v.Kind()
		if k >= reflect.Chan && k <= reflect.Slice && v.IsNil() {
			return true
		}

		return false
	}

	if reflect.DeepEqual(e, g) {
		return true
	}

	if c.intEqual(e, g) {
		return true
	}

	if c.floatingEqual(e, g) {
		return true
	}

	return false
}

func (c *C) contains(iter, v interface{}) (found, ok bool) {
	ok = true

	iterV := reflect.ValueOf(iter)
	vv := reflect.ValueOf(v)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()

	if iterV.Kind() == reflect.Map {
		keys := iterV.MapKeys()
		for _, k := range keys {
			if c.equal(k.Interface(), v) {
				found = true
				return
			}
		}

		return
	}

	if iterV.Kind() == reflect.Slice || iterV.Kind() == reflect.Array {
		if vv.Kind() == reflect.Slice || vv.Kind() == reflect.Array {
			at := 0

			for i := 0; i < iterV.Len(); i++ {
				if at == vv.Len() {
					break
				}

				if c.equal(iterV.Index(i).Interface(), vv.Index(at).Interface()) {
					at++
				} else {
					at = 0
				}
			}

			found = at == vv.Len()
		} else {
			for i := 0; i < iterV.Len(); i++ {
				if c.equal(iterV.Index(i).Interface(), v) {
					found = true
					break
				}
			}
		}

		return
	}

	if vv.Kind() == reflect.String {
		found = strings.Contains(iterV.String(), vv.String())
		return
	}

	ok = false
	return
}

func (c *C) fail(msg ...interface{}) {
	c.Errorf("%s\t%s: %s",
		stack.ClearTestCaller(),
		callerInfo(),
		format(msg...))
}

// True checks that the given bool is true. Returns the value of the bool.
func (c *C) True(cond bool, msg ...interface{}) bool {
	if !cond {
		c.fail("%s\n"+
			"Bool check failed: expected true",
			format(msg...))
	}

	return cond
}

// MustTrue is like True, except it panics on failure.
func (c *C) MustTrue(cond bool, msg ...interface{}) {
	if !c.True(cond, msg...) {
		c.FailNow()
	}
}

// False checks that the given bool is false. Returns the value opposite value
// of the bool.
func (c *C) False(cond bool, msg ...interface{}) bool {
	if cond {
		c.fail("%s\n"+
			"Bool check failed: expected false",
			format(msg...))
	}

	return !cond
}

// MustFalse is like False, except it panics on failure.
func (c *C) MustFalse(cond bool, msg ...interface{}) {
	if !c.False(cond, msg...) {
		c.FailNow()
	}
}

// Equal compares to things, ensuring that they are equal to each other. `e`
// is the expected value; `g` is the value you got somewhere else. Returns
// true if they not equal, false otherwise.
//
// Equal takes special care of floating point numbers, ensuring that any
// precision loss doesn't affect their equality.
//
// If `e` is nil, `g` will be checked for nil, and if it's an interface, its
// value will be checked for nil. Keep in mind that, for interfaces, this is
// _not_ a strict `g == nil` comparison.
func (c *C) Equal(e, g interface{}, msg ...interface{}) bool {
	if !c.equal(e, g) {
		c.fail("%s\n"+
			"Expected: %+v\n"+
			"       == %+v",
			format(msg...),
			e,
			g)
		return false
	}

	return true
}

// MustEqual is like Equal, except it panics on failure.
func (c *C) MustEqual(e, g interface{}, msg ...interface{}) {
	if !c.Equal(e, g, msg...) {
		c.FailNow()
	}
}

// NotEqual compares to things, ensuring that they do not equal each other.
// Returns true if they are not equal, false otherwise.
//
// NotEqual takes special care of floating point numbers, ensuring that any
// precision loss doesn't affect their equality.
func (c *C) NotEqual(e, g interface{}, msg ...interface{}) bool {
	if c.equal(e, g) {
		c.fail("%s\n"+
			"Expected %+v\n"+
			"      != %+v",
			format(msg...),
			e,
			g)
		return false
	}

	return true
}

// MustNotEqual is like NotEqual, except it panics on failure.
func (c *C) MustNotEqual(e, g interface{}, msg ...interface{}) {
	if !c.NotEqual(e, g, msg...) {
		c.FailNow()
	}
}

// Len checks that the length of the given v is l. Returns true if equal,
// false otherwise.
func (c *C) Len(v interface{}, l int, msg ...interface{}) (eq bool) {
	defer func() {
		if e := recover(); e != nil {
			eq = false
			c.fail("%s\n"+
				"%+v is not iterable, cannot check length",
				format(msg...),
				v)
		}
	}()

	vv := reflect.ValueOf(v)
	eq = c.Equal(l, vv.Len(), msg...)
	return
}

// MustLen is like Len, except it panics on failure.
func (c *C) MustLen(v interface{}, l int, msg ...interface{}) {
	if !c.Len(v, l, msg...) {
		c.FailNow()
	}
}

// LenNot checks that the length of the given v is not l. Returns true if not
// equal, false otherwise.
func (c *C) LenNot(v interface{}, l int, msg ...interface{}) (eq bool) {
	defer func() {
		if e := recover(); e != nil {
			eq = false
			c.fail("%s\n"+
				"%+v is not iterable, cannot check length",
				format(msg...),
				v)
		}
	}()

	vv := reflect.ValueOf(v)
	eq = c.NotEqual(vv.Len(), l, msg...)
	return
}

// MustLenNot is like LenNot, except it panics on failure.
func (c *C) MustLenNot(v interface{}, l int, msg ...interface{}) {
	if !c.LenNot(v, l, msg...) {
		c.FailNow()
	}
}

// Contains checks that iter contains v. Returns true if it does, false otherwise.
//
// The following checks are done:
//    0) If a map, checks if the map contains key v.
//    1) If iter is a slice/array and v is a slice/array, checks to see if v
//       is a subset of iter.
//    2) If iter is a slice/array and v is not, checks if any element in iter
//       equals v.
//    3) If iter is a string, falls back to strings.Contains.
func (c *C) Contains(iter, v interface{}, msg ...interface{}) bool {
	found, ok := c.contains(iter, v)

	if !ok {
		c.fail("%s\n"+
			"%+v is not iterable; contain check failed",
			format(msg...),
			v)
		return false
	}

	if !found {
		c.fail("%s\n"+
			"%+v does not contain %+v",
			format(msg...),
			iter,
			v)
		return false
	}

	return true
}

// MustContain is like Contains, except it panics on failure.
func (c *C) MustContain(iter, v interface{}, msg ...interface{}) {
	if !c.Contains(iter, v, msg...) {
		c.FailNow()
	}
}

// NotContains checks that v does not contain c. Returns true if it does,
// false otherwise.
func (c *C) NotContains(iter, v interface{}, msg ...interface{}) bool {
	found, ok := c.contains(iter, v)

	if !ok {
		c.fail("%s\n"+
			"%+v is not iterable; contain check failed",
			format(msg...),
			v)
		return false
	}

	if found {
		c.fail("%s\n"+
			"%+v contains %+v",
			format(msg...),
			iter,
			v)
		return false
	}

	return true
}

// MustNotContain is like NotContains, except it panics on failure.
func (c *C) MustNotContain(iter, v interface{}, msg ...interface{}) {
	if !c.NotContains(iter, v, msg...) {
		c.FailNow()
	}
}

// Is ensures that g is the same type as e. Returns true if they are the same
// type, false otherwise.
func (c *C) Is(e, g interface{}, msg ...interface{}) bool {
	te := reflect.TypeOf(e)
	tg := reflect.TypeOf(g)

	if !c.equal(te, tg) {
		c.fail("%s\n"+
			"Expected type: %s.%s\n"+
			"            == %s.%s",
			format(msg...),
			te.PkgPath(), te.Name(),
			tg.PkgPath(), tg.Name())
		return false
	}

	return true
}

// MustBe is like Is, except it panics on failure.
func (c *C) MustBe(e, g interface{}, msg ...interface{}) {
	if !c.Is(e, g, msg...) {
		c.FailNow()
	}
}

// IsNot ensures that g is not the same type as e. Returns true if they are
// not the same type, false otherwise.
func (c *C) IsNot(e, g interface{}, msg ...interface{}) bool {
	te := reflect.TypeOf(e)
	tg := reflect.TypeOf(g)

	if c.equal(te, tg) {
		c.fail("%s\n"+
			"Expected type: %s.%s\n"+
			"            != %s.%s",
			format(msg...),
			te.PkgPath(), te.Name(),
			tg.PkgPath(), tg.Name())
		return false
	}

	return true
}

// MustNotBe is like IsNot, except it panics on failure.
func (c *C) MustNotBe(e, g interface{}, msg ...interface{}) {
	if !c.IsNot(e, g, msg...) {
		c.FailNow()
	}
}

// Error ensures that an error is not nil. Returns true if an error was
// received, false otherwise.
func (c *C) Error(err error, msg ...interface{}) bool {
	if err == nil {
		c.fail("%s\n"+
			"Expected an error, got nil",
			format(msg...))
		return false
	}

	return true
}

// MustError is like Error, except it panics on failure.
func (c *C) MustError(err error, msg ...interface{}) {
	if !c.Error(err, msg...) {
		c.FailNow()
	}
}

// NotError ensures that an error is nil. Returns true if no error was found,
// false otherwise.
func (c *C) NotError(err error, msg ...interface{}) bool {
	if err != nil {
		c.fail("%s\n"+
			"Expected no error, got: %s",
			format(msg...),
			err)
		return false
	}

	return true
}

// MustNotError is like NotError, except it panics on failure.
func (c *C) MustNotError(err error, msg ...interface{}) {
	if !c.NotError(err, msg...) {
		c.FailNow()
	}
}

// Panics ensures that the given function panics
func (c *C) Panics(fn func(), msg ...interface{}) (ok bool) {
	defer func() {
		if r := recover(); r == nil {
			c.fail("%s\n"+
				"Expected fn to panic; it did not.",
				format(msg...))
		} else {
			ok = true
		}
	}()

	fn()
	return
}

// MustPanic is like Panic, except it panics on failure.
func (c *C) MustPanic(fn func(), msg ...interface{}) {
	if !c.Panics(fn, msg...) {
		c.FailNow()
	}
}

// NotPanic ensures that the given function does not panic
func (c *C) NotPanic(fn func(), msg ...interface{}) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			c.fail("%s\n"+
				"Expected fn not to panic; got panic with: %+v",
				format(msg...),
				r)
		} else {
			ok = true
		}
	}()

	fn()
	return
}

// MustNotPanic is like NotPanic, except it panics on failure.
func (c *C) MustNotPanic(fn func(), msg ...interface{}) {
	if !c.NotPanic(fn, msg...) {
		c.FailNow()
	}
}

// Until polls for the given function for the given amount of time. If in that
// time the function did not return true, the test fails immediately.
func (c *C) Until(wait time.Duration, fn func() bool, msg ...interface{}) {
	sleep := wait / 1000
	for i := 0; i < 1000; i++ {
		if fn() {
			return
		}

		time.Sleep(sleep)
	}

	c.fail("%s\n"+
		"Waiting for condition failed",
		format(msg...))
	c.FailNow()
}
