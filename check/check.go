// Package check provides dead-simple assertions and utilities for testing.
//
// All tests created with check.New() run in parallel, so be warned.
package check

import (
	"fmt"
	"testing"
)

// C is like *testing.T/*testing.B, but with more fun
type C struct {
	testing.TB
	Asserter
	Must Asserter
}

// New creates a new C and marks this test as parallel
func New(tb testing.TB) *C {
	if t, ok := tb.(*testing.T); ok {
		t.Parallel()
	}

	return &C{
		TB:       tb,
		Asserter: newNoopAssert(tb),
		Must:     newMustAssert(tb),
	}
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

// Run is the equivalent of testing.{T,B}.Run()
func (c *C) Run(name string, fn func(*C)) bool {
	switch tb := c.TB.(type) {
	case *testing.T:
		return tb.Run(name, func(t *testing.T) {
			fn(New(t))
		})

	case *testing.B:
		return tb.Run(name, func(b *testing.B) {
			fn(New(b))
		})

	default:
		panic(fmt.Errorf("unsupported testing.TB: %T", tb))
	}
}
