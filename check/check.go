// Package check provides dead-simple assertions and utilities for testing.
//
// All tests created with check.New() run in parallel, so be warned.
package check

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

// C is like *testing.T/*testing.B, but with more fun
type C struct {
	testing.TB
	Asserter
	Must Asserter
	*c
	name string
}

const unknownFunc = "???:1"

// New creates a new C and marks this test as parallel
func New(tb testing.TB) *C {
	c := c{
		path: getCallerPath(),
	}

	return c.newChild(tb)
}

// Name gets the name of this test
func (c *C) Name() string {
	return c.name
}

// T provides access to the underlying *testing.T. If C was not instantiated
// with a *testing.T, this panics.
func (c *C) T() *testing.T {
	return c.TB.(*testing.T)
}

// B provides access to the underlying *testing.B. If C was not instantiated
// with a *testing.B, this panics.
func (c *C) B() *testing.B {
	return c.TB.(*testing.B)
}

// FS gets the shared FS for this test tree. If you want an FS isolated to
// your test, call NewFS().
//
// For this to work, there must be a "test_data" directory somewhere in the
// parent directories of your test. All test files are put into this directory
// at runtime, and they're cleaned up on test success. All tests may safely
// share the same directory.
//
// Be sure to call cleanup() when you're done with the FS. If this has been
// called multiple times, cleanup() must be called the same number of times to
// ensure that the dir is cleaned up.
func (c *C) FS() (fs *FS, cleanup func()) {
	c.fsMtx.Lock()
	defer c.fsMtx.Unlock()

	if c.fs == nil {
		c.fs, cleanup = newFS(c, 0)
		fs = c.fs
		return
	}

	return c.fs.ref()
}

// NewFS creates a new, isolated FS.
//
// Be sure to call cleanup() when you're done with the FS.
func (c *C) NewFS() (fs *FS, cleanup func()) {
	c.fsMtx.Lock()
	defer c.fsMtx.Unlock()

	c.fsI++
	return newFS(c, c.fsI)
}

// Run is the equivalent of testing.{T,B}.Run()
func (c *C) Run(name string, fn func(*C)) bool {
	switch tb := c.TB.(type) {
	case *testing.T:
		return tb.Run(name, func(t *testing.T) {
			fn(c.newChild(t))
		})

	default:
		panic(fmt.Errorf("unsupported testing.TB: %T", tb))
	}
}

// c is shared amongst all tests in a test tree
type c struct {
	path string

	fsMtx sync.Mutex
	fsI   int
	fs    *FS
}

func (c *c) newChild(tb testing.TB) *C {
	if t, ok := tb.(*testing.T); ok {
		t.Parallel()
	}

	cc := &C{
		TB:       tb,
		Asserter: newNoopAssert(tb),
		Must:     newMustAssert(tb),
		c:        c,
		name:     getTestName(tb),
	}

	return cc
}

func getTestName(tb testing.TB) string {
	switch v := tb.(type) {
	case *testing.T, *testing.B:
		rv := reflect.Indirect(reflect.ValueOf(v))
		return rv.FieldByName("name").String()

	default:
		panic(fmt.Errorf("unsupported testing.TB: %T", tb))
	}
}

func getCallerPath() string {
	pc, _, _, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	file, _ := fn.FileLine(pc)
	return filepath.Dir(file)
}
