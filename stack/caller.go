// Package stack provides some utilities for dealing with the call stack.
package stack

import (
	"fmt"
	"path/filepath"
	"runtime"
)

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
