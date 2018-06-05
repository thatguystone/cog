package check_test

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func Example_check() {
	// Typically you would pass in your *testing.T or *testing.B here
	c := check.New(new(testing.B))

	// These are just a few of the provided functions. Check out the full
	// documentation for everything.

	c.Equal(1, 1, "the universe is falling apart")
	c.NotEqual(1, 2, "those can't be equal!")

	panics := func() {
		panic("i get nervous sometimes")
	}
	c.Panics(panics, "this should always panic")
}
