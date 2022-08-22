package check_test

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func Example_check() {
	// Typically you would pass in your *testing.T
	c := check.NewT(new(testing.T))

	// These are just a few of the provided functions. Check out the full
	// documentation for everything.

	c.Equal(1, 1)
	c.NotEqualf(1, 2, "some format %s", "string")

	c.Panics(func() {
		panic("i get nervous sometimes")
	})
}
