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

// Caller formats the caller at `depth` like "filename:lineno"
func Caller(depth int) (c string) {
	c = "???:1"

	_, file, line, ok := runtime.Caller(depth + 1)
	if ok {
		c = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	return c
}
