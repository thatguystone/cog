package check

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/iheartradio/cog/cfs"
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

// FileExists checks that the given file exists
func (fs *FS) FileExists(path string) {
	exists, err := cfs.FileExists(fs.Path(path))
	fs.c.MustNotError(err)
	fs.c.True(exists, "%s does not exist", path)
}

// FileNotExists checks that the given file does not exist
func (fs *FS) FileNotExists(path string) {
	exists, err := cfs.FileExists(fs.Path(path))
	fs.c.MustNotError(err)
	fs.c.False(exists, "%s exists when it shouldn't", path)
}

// DirExists checks that the given directory exists
func (fs *FS) DirExists(path string) {
	exists, err := cfs.DirExists(fs.Path(path))
	fs.c.MustNotError(err)
	fs.c.True(exists, "%s does not exist", path)
}

// DirNotExists checks that the given directory exists
func (fs *FS) DirNotExists(path string) {
	exists, err := cfs.DirExists(fs.Path(path))
	fs.c.MustNotError(err)
	fs.c.False(exists, "%s exists when it shouldn't", path)
}

// WriteFile writes the given contents to the given path in the test's data
// dir, creating everything as necessary.
func (fs *FS) WriteFile(path string, contents []byte) {
	err := cfs.Write(fs.Path(path), contents)
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

// ContentsEqual checks that the contents of the given file exactly equal the
// given byte slice.
func (fs *FS) ContentsEqual(path string, b []byte) {
	a := fs.ReadFile(path)
	fs.c.Equal(b, a)
}

// SContentsEqual is like ContentsEqual, but with strings.
func (fs *FS) SContentsEqual(path string, b string) {
	a := fs.SReadFile(path)
	fs.c.Equal(b, a)
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
		dDir, err := cfs.FindDirInParents(DataDir)
		fs.c.MustNotError(err)

		tDir, name := GetTestDetails()
		relPath, err := filepath.Rel(dDir, tDir)
		fs.c.MustNotError(err)

		path := filepath.Clean(filepath.Join("/", relPath))

		fs.dataDir = filepath.Join(dDir, path, name)

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

// DumpTree walks the tree at the given path and writes each file to the test
// log
func (fs *FS) DumpTree(path string) {
	fs.c.Logf("Tree at `%s`:", path)

	rootPath := fs.Path(path)
	filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			fs.c.MustNotError(err)

			rel, err := filepath.Rel(rootPath, path)
			fs.c.MustNotError(err)

			fs.c.Logf("\t%s", rel)
			return nil
		})
}
