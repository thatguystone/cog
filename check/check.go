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

var (
	// Used to ensure that t.Parallel() is only called once
	mtx          sync.Mutex
	parallelized = map[string]struct{}{}
)

// New creates a new C and marks this test as parallel
func New(tb testing.TB) *C {
	if t, ok := tb.(*testing.T); ok {
		name := GetTestName()

		alreadyParallel := func() bool {
			mtx.Lock()
			defer mtx.Unlock()

			_, ok := parallelized[name]
			if !ok {
				parallelized[name] = struct{}{}
			}

			return ok
		}()

		if !alreadyParallel {
			t.Parallel()
		}
	}

	c := &C{
		TB: tb,
	}

	c.FS.c = c

	return c
}

// B provides access to the underlying *testing.B. If A was not instantiated
// with a *testing.B, this panics.
func (c *C) B() *testing.B {
	return c.TB.(*testing.B)
}

// T provides access to the underlying *testing.T. If A was not instantiated
// with a *testing.T, this panics.
func (c *C) T() *testing.T {
	return c.TB.(*testing.T)
}

// GetTestName gets the name of the current test.
func GetTestName() string {
	name := "???:1"

	ok := true
	for i := 0; ok; i++ {
		var pc uintptr

		pc, _, _, ok = runtime.Caller(i)
		if ok {
			fn := runtime.FuncForPC(pc)
			fnName := filepath.Ext(fn.Name())

			isTest := strings.Contains(fnName, ".Test") ||
				strings.Contains(fnName, ".Benchmark") ||
				strings.Contains(fnName, ".Example")
			if isTest {
				name = fnName[1:]
				break
			}
		}
	}

	return name
}
