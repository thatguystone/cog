package check

import (
	"path/filepath"
	"sync"

	"github.com/iheartradio/cog/cfs"
)

var (
	// FixtureDir is the directory that all of your fixtures live in. This
	// directory is located by walking up the tree, starting at the CWD until a
	// directory with the given name is found. Typically, this will be
	// "test_fixtures", but you may choose your own name.
	FixtureDir = "test_fixtures"

	fixtureOnce sync.Once
	fixturePath = ""
)

//gocovr:skip-file

// Fixture gets the path to the fixture described by `parts`. If the FixtureDir
// is not found, this panics.
//
// The fixture directory is found by climbing the FS tree until a directory
// containing the directory defined by the constant FixtureDir is found. If
// found, the absolute path to that directory is joined with the given parts is
// returned.
func Fixture(parts ...string) string {
	fixtureOnce.Do(func() {
		path, err := cfs.FindDirInParents(FixtureDir)
		if err != nil {
			panic(err)
		}

		fixturePath = path
	})

	return filepath.Join(append([]string{fixturePath}, parts...)...)
}
