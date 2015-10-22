package check

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/thatguystone/cog/cfs"
)

// FS provides access to a test's directory
type FS struct {
	c           *C
	dataDirOnce sync.Once
	dataDir     string
}

const dirPerms = 0700

// DataDir is the name of the directory where test-specific files live; that is,
// files created specifically for a single test while running the test.
const DataDir = "test_data"

// Path gets the absolute path of the test's data dir joined with the parts. The
// parent dirs are created, assuming that the last part is the file.
func (fs *FS) Path(parts ...string) string {
	parts = append([]string{fs.GetDataDir()}, parts...)
	path := filepath.Join(parts...)

	parent := filepath.Join(path, "..")
	err := os.MkdirAll(parent, dirPerms)
	fs.c.MustNotError(err)

	return path
}

// WriteFile writes the given contents to the given path in the test's data
// dir, creating everything as necessary.
func (fs *FS) WriteFile(path string, contents []byte) {
	path = fs.Path(path)

	err := ioutil.WriteFile(path, contents, 0600)
	fs.c.MustNotError(err)
}

// SWriteFile is like WriteFile, but it works on strings
func (fs *FS) SWriteFile(path, contents string) {
	fs.WriteFile(path, []byte(contents))
}

// ReadFile reads the contents of the file at given path in the test's data dir.
func (fs *FS) ReadFile(path string) []byte {
	path = fs.Path(path)

	b, err := ioutil.ReadFile(path)
	fs.c.MustNotError(err)

	return b
}

// SReadFile is like ReadFile, but it works on strings.
func (fs *FS) SReadFile(path string) string {
	return string(fs.ReadFile(path))
}

// GetDataDir gets the test's data directory. On the first call, this also
// clears out any data directory that existed previously, giving your test a
// clean space to run.
//
// If the DataDir is not found, this panics.
func (fs *FS) GetDataDir() string {
	if fs.dataDir != "" {
		return fs.dataDir
	}

	fs.dataDirOnce.Do(func() {
		path, err := cfs.FindDirInParents(DataDir)
		fs.c.MustNotError(err)

		fs.dataDir = filepath.Join(path, GetTestName())

		// Since this is only called once per test, clear it on first access
		fs.CleanDataDir()
		os.Mkdir(fs.dataDir, dirPerms)
	})

	return fs.dataDir
}

// CleanDataDir wipes out this test's data directory
func (fs *FS) CleanDataDir() {
	os.RemoveAll(fs.GetDataDir())
}
