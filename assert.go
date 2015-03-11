// Package assert provides dead-simple assertions for testing.
package assert

import (
	"fmt"
	"math"
	"math/big"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Tester provides the necessary testing functions for assertions
type Tester interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// A is like *testing.T/*testing.B, but it provides assertions
type A struct {
	Tester
}

// From is a workaround for `go vet`'s complaining about a composite literal
// using unkeyed fields.
func From(t Tester) A {
	return A{t}
}

func (a A) format(msg ...interface{}) string {
	if msg == nil || len(msg) == 0 {
		return ""
	} else if len(msg) == 1 {
		return msg[0].(string)
	} else {
		return fmt.Sprintf(msg[0].(string), msg[1:]...)
	}
}

func (a A) clearInternalCaller() string {
	l := len("???:1")
	_, file, line, ok := runtime.Caller(1)

	if ok {
		// +8 for leading tab
		l = len(fmt.Sprintf("%s:%d: ", path.Base(file), line)) + 8
	}

	return strings.Repeat(" ", l)
}

func (a A) callerInfo() string {
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
			} else {
				return fmt.Sprintf("%s (from %s:%d)", in, file, line)
			}
		} else {
			if in == "" {
				in = fmt.Sprintf("%s:%d", file, line)
			}
		}
	}

	return in
}

func (a A) getInt(e interface{}) (*big.Int, bool) {
	i := big.NewInt(0)

	// Maybe I'm just being stupid today, but go isn't letting me do any
	// simple type casting.
	switch v := e.(type) {
	case int:
		i.SetInt64(int64(v))
	case int8:
		i.SetInt64(int64(v))
	case int16:
		i.SetInt64(int64(v))
	case int32:
		i.SetInt64(int64(v))
	case int64:
		i.SetInt64(int64(v))

	case uint:
		i.SetUint64(uint64(v))
	case uint8:
		i.SetUint64(uint64(v))
	case uint16:
		i.SetUint64(uint64(v))
	case uint32:
		i.SetUint64(uint64(v))
	case uint64:
		i.SetUint64(uint64(v))

	default:
		return nil, false
	}

	return i, true
}

func (a A) intEqual(e, g interface{}) bool {
	ex, ok1 := a.getInt(e)
	gx, ok2 := a.getInt(g)
	if !ok1 || !ok2 {
		return false
	}

	return ex.Cmp(gx) == 0
}

func (a A) floatingEqual(e, g interface{}) bool {
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

func (a A) equal(e, g interface{}) bool {
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

	if a.intEqual(e, g) {
		return true
	}

	if a.floatingEqual(e, g) {
		return true
	}

	return false
}

func (a A) contains(v, c interface{}) (found, ok bool) {
	ok = true

	vv := reflect.ValueOf(v)
	cv := reflect.ValueOf(c)
	defer func() {
		if e := recover(); e != nil {
			ok = false
		}
	}()

	if reflect.TypeOf(v).Kind() == reflect.String {
		found = strings.Contains(vv.String(), cv.String())
		return
	}

	for i := 0; i < vv.Len(); i++ {
		if a.equal(vv.Index(i).Interface(), c) {
			found = true
			return
		}
	}

	return
}

func (a A) fail(msg ...interface{}) {
	a.Errorf("\r%s\r\t%s: %s",
		a.clearInternalCaller(),
		a.callerInfo(),
		a.format(msg...))
}

// True checks that the given bool is true. Returns the value of the bool.
func (a A) True(cond bool, msg ...interface{}) bool {
	if !cond {
		a.fail("%s\n"+
			"Bool check failed: expected true",
			a.format(msg...))
	}

	return cond
}

// MustTrue is like True, except it panics on failure.
func (a A) MustTrue(cond bool, msg ...interface{}) {
	if !a.True(cond, msg...) {
		a.FailNow()
	}
}

// False checks that the given bool is false. Returns the value opposite value
// of the bool.
func (a A) False(cond bool, msg ...interface{}) bool {
	if cond {
		a.fail("%s\n"+
			"Bool check failed: expected false",
			a.format(msg...))
	}

	return !cond
}

// MustFalse is like False, except it panics on failure.
func (a A) MustFalse(cond bool, msg ...interface{}) {
	if !a.False(cond, msg...) {
		a.FailNow()
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
func (a A) Equal(e, g interface{}, msg ...interface{}) bool {
	if !a.equal(e, g) {
		a.fail("%s\n"+
			"Expected: %+v\n"+
			"       == %+v",
			a.format(msg...),
			e,
			g)
		return false
	}

	return true
}

// MustEqual is like Equal, except it panics on failure.
func (a A) MustEqual(e, g interface{}, msg ...interface{}) {
	if !a.Equal(e, g, msg...) {
		a.FailNow()
	}
}

// NotEqual compares to things, ensuring that they do not equal each other.
// Returns true if they are not equal, false otherwise.
//
// NotEqual takes special care of floating point numbers, ensuring that any
// precision loss doesn't affect their equality.
func (a A) NotEqual(e, g interface{}, msg ...interface{}) bool {
	if a.equal(e, g) {
		a.fail("%s\n"+
			"Expected %+v\n"+
			"      != %+v",
			a.format(msg...),
			e,
			g)
		return false
	}

	return true
}

// MustNotEqual is like NotEqual, except it panics on failure.
func (a A) MustNotEqual(e, g interface{}, msg ...interface{}) {
	if !a.NotEqual(e, g, msg...) {
		a.FailNow()
	}
}

// Len checks that the length of the given v is l. Returns true if equal,
// false otherwise.
func (a A) Len(v, l interface{}, msg ...interface{}) (eq bool) {
	defer func() {
		if e := recover(); e != nil {
			eq = false
			a.fail("%s\n"+
				"%+v is not iterable, cannot check length",
				a.format(msg...),
				v)
		}
	}()

	vv := reflect.ValueOf(v)
	eq = a.Equal(vv.Len(), l, msg...)
	return
}

// MustLen is like Len, except it panics on failure.
func (a A) MustLen(v, l interface{}, msg ...interface{}) {
	if !a.Len(v, l, msg...) {
		a.FailNow()
	}
}

// LenNot checks that the length of the given v is not l. Returns true if not
// equal, false otherwise.
func (a A) LenNot(v, l interface{}, msg ...interface{}) (eq bool) {
	defer func() {
		if e := recover(); e != nil {
			eq = false
			a.fail("%s\n"+
				"%+v is not iterable, cannot check length",
				a.format(msg...),
				v)
		}
	}()

	vv := reflect.ValueOf(v)
	eq = a.NotEqual(vv.Len(), l, msg...)
	return
}

// MustLenNot is like LenNot, except it panics on failure.
func (a A) MustLenNot(v, l interface{}, msg ...interface{}) {
	if !a.LenNot(v, l, msg...) {
		a.FailNow()
	}
}

// Contains checks that v contains c. Returns true if it does, false otherwise.
func (a A) Contains(v, c interface{}, msg ...interface{}) bool {
	found, ok := a.contains(v, c)

	if !ok {
		a.fail("%s\n"+
			"%+v is not iterable; contain check failed",
			a.format(msg...),
			v)
		return false
	}

	if !found {
		a.fail("%s\n"+
			"%+v does not contain %+v",
			a.format(msg...),
			v,
			c)
		return false
	}

	return true
}

// MustContain is like Contains, except it panics on failure.
func (a A) MustContain(v, c interface{}, msg ...interface{}) {
	if !a.Contains(v, c, msg...) {
		a.FailNow()
	}
}

// NotContains checks that v does not contain c. Returns true if it does,
// false otherwise.
func (a A) NotContains(v, c interface{}, msg ...interface{}) bool {
	found, ok := a.contains(v, c)

	if !ok {
		a.fail("%s\n"+
			"%+v is not iterable; contain check failed",
			a.format(msg...),
			v)
		return false
	}

	if found {
		a.fail("%s\n"+
			"%+v contains %+v",
			a.format(msg...),
			v,
			c)
		return false
	}

	return true
}

// MustNotContain is like NotContains, except it panics on failure.
func (a A) MustNotContain(v, c interface{}, msg ...interface{}) {
	if !a.NotContains(v, c, msg...) {
		a.FailNow()
	}
}

// Is ensures that g is the same type a e. Returns true if they are the same
// type, false otherwise.
func (a A) Is(e, g interface{}, msg ...interface{}) bool {
	te := reflect.TypeOf(e)
	tg := reflect.TypeOf(g)

	if !a.equal(te, tg) {
		a.fail("%s\n"+
			"Expected type: %s.%s\n"+
			"            == %s.%s",
			a.format(msg...),
			te.PkgPath(), te.Name(),
			tg.PkgPath(), tg.Name())
		return false
	}

	return true
}

// MustBe is like Is, except it panics on failure.
func (a A) MustBe(e, g interface{}, msg ...interface{}) {
	if !a.Is(e, g, msg...) {
		a.FailNow()
	}
}

// IsNot ensures that g is not the same type as e. Returns true if they are
// not the same type, false otherwise.
func (a A) IsNot(e, g interface{}, msg ...interface{}) bool {
	te := reflect.TypeOf(e)
	tg := reflect.TypeOf(g)

	if a.equal(te, tg) {
		a.fail("%s\n"+
			"Expected type: %s.%s\n"+
			"            != %s.%s",
			a.format(msg...),
			te.PkgPath(), te.Name(),
			tg.PkgPath(), tg.Name())
		return false
	}

	return true
}

// MustNotBe is like IsNot, except it panics on failure.
func (a A) MustNotBe(e, g interface{}, msg ...interface{}) {
	if !a.IsNot(e, g, msg...) {
		a.FailNow()
	}
}

// Error ensures that an error is not nil. Returns true if an error was
// received, false otherwise.
func (a A) Error(err error, msg ...interface{}) bool {
	if err == nil {
		a.fail("%s\n"+
			"Expected an error, got nil",
			a.format(msg...))
		return false
	}

	return true
}

// MustError is like Error, except it panics on failure.
func (a A) MustError(err error, msg ...interface{}) {
	if !a.Error(err, msg...) {
		a.FailNow()
	}
}

// NotError ensures that an error is nil. Returns true if no error was found,
// false otherwise.
func (a A) NotError(err error, msg ...interface{}) bool {
	if err != nil {
		a.fail("%s\n"+
			"Expected no error, got: %s",
			a.format(msg...),
			err)
		return false
	}

	return true
}

// MustNotError is like NotError, except it panics on failure.
func (a A) MustNotError(err error, msg ...interface{}) {
	if !a.NotError(err, msg...) {
		a.FailNow()
	}
}

// Panic ensures that the given function panics
func (a A) Panic(fn func(), msg ...interface{}) (ok bool) {
	defer func() {
		if r := recover(); r == nil {
			a.fail("%s\n"+
				"Expected fn to panic; it did not.",
				a.format(msg...))
		} else {
			ok = true
		}
	}()

	fn()
	return
}

// MustPanic is like Panic, except it panics on failure.
func (a A) MustPanic(fn func(), msg ...interface{}) {
	if !a.Panic(fn, msg...) {
		a.FailNow()
	}
}

// NotPanic ensures that the given function does not panic
func (a A) NotPanic(fn func(), msg ...interface{}) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			a.fail("%s\n"+
				"Expected fn not to panic; got panic with: %+v",
				a.format(msg...),
				r)
		} else {
			ok = true
		}
	}()

	fn()
	return
}

// MustNotPanic is like NotPanic, except it panics on failure.
func (a A) MustNotPanic(fn func(), msg ...interface{}) {
	if !a.NotPanic(fn, msg...) {
		a.FailNow()
	}
}

// B provides access to the underlying *testing.B. If A was not instantiated
// with a *testing.B, this panics.
func (a A) B() *testing.B {
	return a.Tester.(*testing.B)
}

// T provides access to the underlying *testing.T. If A was not instantiated
// with a *testing.T, this panics.
func (a A) T() *testing.T {
	return a.Tester.(*testing.T)
}
