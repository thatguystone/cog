package check

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureBasic(t *testing.T) {
	c := New(t)

	err := os.Mkdir("test_fixtures", dirPerms)
	c.MustNotError(err)
	defer os.RemoveAll("test_fixtures")

	abs, err := filepath.Abs("test_fixtures/test")
	c.MustNotError(err)
	c.Equal(abs, Fixture("test"))
}
