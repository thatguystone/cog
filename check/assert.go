package check

import (
	"fmt"
	"math"
	"math/big"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/thatguystone/cog/stack"
	"github.com/thatguystone/cog/stringc"
)

// Can't really get 100% coverage on this file: doing so causes the tests to
// fail.
//gocovr:skip-file

// An Asserter provides test assertions. Each assertion returns if the
// assertion succeeded.
type Asserter interface {
	// True checks that the given bool is true.
	True(cond bool, msg ...interface{}) bool

	// False checks that the given bool is false.
	False(cond bool, msg ...interface{}) bool

	// Equal compares two things, ensuring that they are equal to each other.
	// `e` is the expected value; `g` is the value you got somewhere else.
	//
	// Equal takes special care of floating point numbers, ensuring that any
	// precision loss doesn't affect their equality.
	//
	// If `e` is nil, `g` will be checked for nil, and if it's an interface,
	// its value will be checked for nil. Keep in mind that, for interfaces,
	// this is _not_ a strict `g == nil` comparison.
	Equal(e, g interface{}, msg ...interface{}) bool

	// NotEqual is the opposite of Equal.
	NotEqual(e, g interface{}, msg ...interface{}) bool

	// Len checks that the length of the given v is l.
	Len(v interface{}, l int, msg ...interface{}) bool

	// NotLen is the opposite of Len.
	NotLen(v interface{}, l int, msg ...interface{}) bool

	// Contains checks that iter contains v.
	//
	// The following checks are done:
	//    0) If a map, checks if the map contains key v.
	//    1) If iter is a slice/array and v is a slice/array, checks to see if
	//       v is a subset of iter.
	//    2) If iter is a slice/array and v is not, checks if any element in
	//       iter equals v.
	//    3) If iter is a string, falls back to strings.Contains.
	Contains(iter, v interface{}, msg ...interface{}) bool

	// NotContains is the opposite of Contains.
	NotContains(iter, v interface{}, msg ...interface{}) bool

	// Is ensures that g is the same type as e.
	Is(e, g interface{}, msg ...interface{}) bool

	// NotIs is the exact opposite of Is.
	NotIs(e, g interface{}, msg ...interface{}) bool

	// Nil ensures that g is nil. This is a strict equality check.
	Nil(g interface{}, msg ...interface{}) bool

	// NotNil is the opposite of Nil.
	NotNil(g interface{}, msg ...interface{}) bool

	// Panics ensures that the given function panics
	Panics(fn func(), msg ...interface{}) bool

	// NotPanics ensures that the given function does not panic
	NotPanics(fn func(), msg ...interface{}) bool

	// Until polls for the given function for the given amount of time. If in
	// that time the function did not return true, the test fails immediately.
	Until(wait time.Duration, fn func() bool, msg ...interface{})
}

type assert struct {
	testing.TB
	onFail func()
}

func newNoopAssert(tb testing.TB) assert {
	return assert{
		TB:     tb,
		onFail: func() {},
	}
}

func newMustAssert(tb testing.TB) assert {
	return assert{
		TB: tb,
		onFail: func() {
			tb.FailNow()
		},
	}
}

func format(msg ...interface{}) string {
	if len(msg) == 0 {
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

func (a assert) getInt(e interface{}) (*big.Int, bool) {
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

func (a assert) intEqual(e, g interface{}) bool {
	ex, ok1 := a.getInt(e)
	gx, ok2 := a.getInt(g)
	if !ok1 || !ok2 {
		return false
	}

	return ex.Cmp(gx) == 0
}

func (a assert) floatingEqual(e, g interface{}) bool {
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

func (a assert) equal(e, g interface{}) bool {
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

func (a assert) contains(iter, v interface{}) (found, ok bool) {
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
			if a.equal(k.Interface(), v) {
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

				if a.equal(iterV.Index(i).Interface(), vv.Index(at).Interface()) {
					at++
				} else {
					at = 0
				}
			}

			found = at == vv.Len()
		} else {
			for i := 0; i < iterV.Len(); i++ {
				if a.equal(iterV.Index(i).Interface(), v) {
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

func (a assert) fail(msg ...interface{}) {
	a.Errorf("%s\t%s: %s",
		stack.ClearTestCaller(),
		callerInfo(),
		format(msg...))
	a.onFail()
}

func (a assert) True(cond bool, msg ...interface{}) bool {
	if !cond {
		a.fail("%s\n"+
			"Bool check failed: expected true",
			format(msg...))
	}

	return cond
}
func (a assert) False(cond bool, msg ...interface{}) bool {
	if cond {
		a.fail("%s\n"+
			"Bool check failed: expected false",
			format(msg...))
	}

	return !cond
}

func (a assert) Equal(e, g interface{}, msg ...interface{}) bool {
	if !a.equal(e, g) {
		diff := diff(e, g)

		if diff != "" {
			diff = "\n\nDiff:\n" + stringc.Indent(diff, spewConfig.Indent)
		}

		e, g := fmtVals(e, g)
		a.fail("%s\n"+
			"Expected: `%+v`\n"+
			"       == `%+v`%s",
			format(msg...),
			e, g, diff)
		return false
	}

	return true
}

func (a assert) NotEqual(e, g interface{}, msg ...interface{}) bool {
	if a.equal(e, g) {
		a.fail("%s\n"+
			"Expected: `%+v`\n"+
			"       != `%+v`%s",
			format(msg...),
			e, g)
		return false
	}

	return true
}

func (a assert) Len(v interface{}, l int, msg ...interface{}) (eq bool) {
	defer func() {
		if e := recover(); e != nil {
			eq = false
			a.fail("%s\n"+
				"%+v is not iterable, cannot check length",
				format(msg...),
				v)
		}
	}()

	vv := reflect.ValueOf(v)
	eq = a.Equal(l, vv.Len(), msg...)
	return
}

func (a assert) NotLen(v interface{}, l int, msg ...interface{}) (eq bool) {
	defer func() {
		if e := recover(); e != nil {
			eq = false
			a.fail("%s\n"+
				"%+v is not iterable, cannot check length",
				format(msg...),
				v)
		}
	}()

	vv := reflect.ValueOf(v)
	eq = a.NotEqual(vv.Len(), l, msg...)
	return
}

func (a assert) Contains(iter, v interface{}, msg ...interface{}) bool {
	found, ok := a.contains(iter, v)

	if !ok {
		a.fail("%s\n"+
			"%+v is not iterable; contain check failed",
			format(msg...),
			v)
		return false
	}

	if !found {
		a.fail("%s\n"+
			"%+v does not contain %+v",
			format(msg...),
			iter,
			v)
		return false
	}

	return true
}

func (a assert) NotContains(iter, v interface{}, msg ...interface{}) bool {
	found, ok := a.contains(iter, v)

	if !ok {
		a.fail("%s\n"+
			"%+v is not iterable; contain check failed",
			format(msg...),
			v)
		return false
	}

	if found {
		a.fail("%s\n"+
			"%+v contains %+v",
			format(msg...),
			iter,
			v)
		return false
	}

	return true
}

func (a assert) Is(e, g interface{}, msg ...interface{}) bool {
	te := reflect.TypeOf(e)
	tg := reflect.TypeOf(g)

	if !a.equal(te, tg) {
		a.fail("%s\n"+
			"Expected type: %s.%s\n"+
			"            == %s.%s",
			format(msg...),
			te.PkgPath(), te.Name(),
			tg.PkgPath(), tg.Name())
		return false
	}

	return true
}

func (a assert) NotIs(e, g interface{}, msg ...interface{}) bool {
	te := reflect.TypeOf(e)
	tg := reflect.TypeOf(g)

	if a.equal(te, tg) {
		a.fail("%s\n"+
			"Expected type: %s.%s\n"+
			"            != %s.%s",
			format(msg...),
			te.PkgPath(), te.Name(),
			tg.PkgPath(), tg.Name())
		return false
	}

	return true
}

func (a assert) Nil(g interface{}, msg ...interface{}) bool {
	if g != nil {
		a.fail("%s\n"+
			"Expected nil, got: `%+v`",
			format(msg...),
			g)
		return false
	}

	return true
}

func (a assert) NotNil(g interface{}, msg ...interface{}) bool {
	if g == nil {
		a.fail("%s\n"+
			"Expected something, got nil",
			format(msg...))
		return false
	}

	return true
}

func (a assert) Panics(fn func(), msg ...interface{}) (ok bool) {
	defer func() {
		if r := recover(); r == nil {
			a.fail("%s\n"+
				"Expected fn to panic; it did not.",
				format(msg...))
		} else {
			ok = true
		}
	}()

	fn()
	return
}

func (a assert) NotPanics(fn func(), msg ...interface{}) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			a.fail("%s\n"+
				"Expected fn not to panic; got: %+v\n%s",
				format(msg...),
				r,
				stringc.Indent(string(debug.Stack()), "\t"))
		} else {
			ok = true
		}
	}()

	fn()
	return
}

func (a assert) Until(timeout time.Duration, fn func() bool, msg ...interface{}) {
	deadline := time.Now().Add(timeout)
	sleep := timeout / 1000
	for time.Now().Before(deadline) {
		if fn() {
			return
		}

		time.Sleep(sleep)
	}

	a.fail("%s\n"+
		"Waiting for condition failed",
		format(msg...))
	a.FailNow()
}
