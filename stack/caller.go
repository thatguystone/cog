// Package stack provides some utilities for dealing with the call stack.
package stack

import (
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// ClearTestCaller returns a string that, when printed to a terminal, clears
// line information typically output by t.Log().
func ClearTestCaller() string {
	l := len("???:1")
	_, file, line, ok := runtime.Caller(1)

	if ok {
		// +8 for leading tab
		l = len(fmt.Sprintf("%s:%d: ", path.Base(file), line)) + 8
	}

	return "\r" + strings.Repeat(" ", l) + "\r"
}

// Caller formats the caller at `depth` like "filename:lineno". Returns
// "???:1" on error.
func Caller(depth int) (c string) {
	c = "???:1"

	_, file, line, ok := runtime.Caller(depth + 1)
	if ok {
		c = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	return c
}

// CallerAbove finds the depth first calling function, above depth, without
// the given prefix (package name, etc). If it can't find such a caller, it
// returns the depth of the stack or the depth of first error.
func CallerAbove(depth int, prefix string) (d int) {
	d = depth + 1

	for {
		d++
		pc, _, _, ok := runtime.Caller(d)
		if !ok {
			break
		}

		name := runtime.FuncForPC(pc).Name()
		if !strings.HasPrefix(name, prefix) {
			break
		}
	}

	d--

	return
}
