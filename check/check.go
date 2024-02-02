package check

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/thatguystone/cog/textwrap"
)

type Error interface {
	Helper()
	Error(args ...any)
}

type Fatal interface {
	Helper()
	Fatal(args ...any)
}

func checkTrue(cond bool) (string, bool) {
	if cond {
		return "", true
	}

	return "Bool check failed: expected true", false
}

func checkFalse(cond bool) (string, bool) {
	if !cond {
		return "", true
	}

	return "Bool check failed: expected false", false
}

func equalMsg(op string, g, e any) string {
	diff := diff(g, e)
	if diff != "" {
		diff = "\n\nDiff:\n" + textwrap.Indent(diff, spewConfig.Indent)
	}

	gf, ef := fmtVals(g, e)
	return fmt.Sprintf(""+
		"Expected: %s\n"+
		"       %s %s%s",
		gf,
		op,
		ef,
		diff,
	)
}

func checkEqual(g, e any) (string, bool) {
	if reflect.DeepEqual(g, e) {
		return "", true
	}

	return equalMsg("==", g, e), false
}

func checkNotEqual(g, e any) (string, bool) {
	if !reflect.DeepEqual(g, e) {
		return "", true
	}

	return equalMsg("!=", g, e), false
}

func checkNil(v any) (string, bool) {
	if v == nil {
		return "", true
	}

	return fmt.Sprintf("Expected nil, got: %+v", v), false
}

func checkNotNil(v any) (string, bool) {
	if v != nil {
		return "", true
	}

	return "Expected something, got nil", false
}

func checkErrIs(err, target error) (string, bool) {
	if errors.Is(err, target) {
		return "", true
	}

	return equalMsg("==", err, target), false
}

func checkErrAs(err error, target any) (string, bool) {
	if errors.As(err, target) {
		return "", true
	}

	return equalMsg("==", err, target), false
}

func containsMsg(container any, what string, el any) string {
	return fmt.Sprintf(
		"%s %+v not found in %+v",
		what,
		el,
		container,
	)
}

func notContainsMsg(container any, what string, el any) string {
	return fmt.Sprintf(
		"%s %+v unexpectedly found in %+v",
		what,
		el,
		container,
	)
}

func hasKey(m, k any) (string, bool) {
	rv := reflect.ValueOf(m)

	if rv.Kind() != reflect.Map {
		msg := fmt.Sprintf("Cannot check non-map %T for key", m)
		return msg, false
	}

	v := rv.MapIndex(reflect.ValueOf(k))
	return "", v != reflect.Value{}
}

func checkHasKey(m, k any) (string, bool) {
	msg, ok := hasKey(m, k)
	if msg != "" {
		return msg, false
	}

	if ok {
		return "", true
	}

	return containsMsg(m, "key", k), false
}

func checkNotHasKey(m, k any) (string, bool) {
	msg, ok := hasKey(m, k)
	if msg != "" {
		return msg, false
	}

	if !ok {
		return "", true
	}

	return notContainsMsg(m, "key", k), false
}

func contains(iter, v any) (msg, what string, ok bool) {
	rv := reflect.ValueOf(iter)

	switch rv.Kind() {
	case reflect.Map:
		for mi := rv.MapRange(); mi.Next(); {
			if reflect.DeepEqual(mi.Value().Interface(), v) {
				return "", "", true
			}
		}

		return "", "value", false

	case reflect.Slice, reflect.Array:
		for i := range rv.Len() {
			if reflect.DeepEqual(rv.Index(i).Interface(), v) {
				return "", "", true
			}
		}

		return "", "value", false

	case reflect.String:
		substr, ok := v.(string)
		if !ok {
			msg := fmt.Sprintf("Cannot search string for non-string: %T(%v)", v, v)
			return msg, "", false
		}

		if strings.Contains(rv.String(), substr) {
			return "", "", true
		}

		return "", "substring", false

	default:
		msg := fmt.Sprintf("Cannot check non-container %T for containment", iter)
		return msg, "", false
	}
}

func checkContains(iter, v any) (string, bool) {
	msg, what, ok := contains(iter, v)
	if msg != "" {
		return msg, false
	}

	if ok {
		return "", true
	}

	return containsMsg(iter, what, v), false
}

func checkNotContains(iter, v any) (string, bool) {
	msg, what, ok := contains(iter, v)
	if msg != "" {
		return msg, false
	}

	if !ok {
		return "", true
	}

	return notContainsMsg(iter, what, v), false
}

func checkPanics(fn func()) (msg string, ok bool) {
	defer func() {
		ok = recover() != nil
	}()

	fn()
	return "Expected fn to panic; it did not.", false
}

func checkNotPanics(fn func()) (msg string, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprintf(
				"Expected fn not to panic; got: %+v\n%s",
				r,
				textwrap.Indent(string(debug.Stack()), "\t"),
			)
		}
	}()

	fn()
	return "", true
}

func checkPanicsWith(recovers any, fn func()) (msg string, ok bool) {
	defer func() {
		r := recover()
		if r == nil {
			msg = "Expected fn to panic; it did not."
			return
		}

		if !reflect.DeepEqual(r, recovers) {
			msg = fmt.Sprintf(
				"%s\n%s",
				equalMsg("==", r, recovers),
				textwrap.Indent(string(debug.Stack()), "\t"),
			)
			return
		}

		ok = true
	}()

	fn()
	return
}

func checkEventuallyTrue(numTries int, fn func(i int) bool) (string, bool) {
	for i := range numTries {
		if fn(i) {
			return "", true
		}
	}

	msg := fmt.Sprintf(
		"Polling for condition failed, gave up after %d tries",
		numTries,
	)
	return msg, false
}

func checkEventuallyNil(numTries int, fn func(i int) error) (string, bool) {
	var err error

	for i := range numTries {
		err = fn(i)
		if err == nil {
			return "", true
		}
	}

	msg := fmt.Sprintf(
		"Func didn't succeed after %d tries, last err: %v",
		numTries,
		err,
	)
	return msg, false
}
