// Package check provides dead-simple assertions and utilities for testing.
//
// All tests created with check.New() run in parallel, so be warned.
package check

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// C is like *testing.T/*testing.B, but with more fun
type C struct {
	testing.TB

	// Access to the test's data directory
	FS FS
}

const unknownFunc = "???:1"

// New creates a new C and marks this test as parallel
func New(tb testing.TB) *C {
	if t, ok := tb.(*testing.T); ok {
		setParallel(t)
	}

	c := &C{
		TB: tb,
	}

	c.FS.c = c

	return c
}

// B provides access to the underlying *testing.B. If C was not instantiated
// with a *testing.B, this panics.
func (c *C) B() *testing.B {
	return c.TB.(*testing.B)
}

// T provides access to the underlying *testing.T. If C was not instantiated
// with a *testing.T, this panics.
func (c *C) T() *testing.T {
	return c.TB.(*testing.T)
}

// Used to ensure that calling New() from multiple goroutines is safe
var (
	mtx          sync.Mutex
	parallelized = map[*testing.T]struct{}{}
)

func setParallel(t *testing.T) {
	mtx.Lock()

	_, alreadyParallel := parallelized[t]
	if !alreadyParallel {
		parallelized[t] = struct{}{}
	}

	mtx.Unlock()

	if !alreadyParallel {
		t.Parallel()
	}
}

// GetTestName gets the name of the current test.
func GetTestName() string {
	_, name := GetTestDetails()
	return name
}

// GetTestDetails gets the file path and name of the test.
func GetTestDetails() (path, name string) {
	name = unknownFunc

	ok := true
	for i := 0; ok; i++ {
		var pc uintptr

		pc, _, _, ok = runtime.Caller(i)
		if ok {
			fn := runtime.FuncForPC(pc)
			file, _ := fn.FileLine(pc)
			fnName := filepath.Ext(fn.Name())

			isTest := strings.Contains(fnName, ".Test") ||
				strings.Contains(fnName, ".Benchmark") ||
				strings.Contains(fnName, ".Example")
			if isTest {
				path = filepath.Dir(file)
				name = fnName[1:]
				break
			}
		}
	}

	return
}
