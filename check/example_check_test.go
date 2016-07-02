package check_test

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Make absolute path relative for example output checking
	wd, _ := os.Getwd()

	// Get a clean directory that is isolated to this test
	fs, cleanup := c.FS()
	defer cleanup() // Be sure to cleanup when done

	rel, _ := filepath.Rel(wd, fs.Path("test_file"))

	// The test data directory is wiped out when all tests pass, so go ahead
	// and make things messy.
	fmt.Println("test-specific file:", rel)

	// Output:
	// test-specific file: ../test_data/check/test_file
}
