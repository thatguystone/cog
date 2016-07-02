package check

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/thatguystone/cog/cfs"
)

// FS provides access to a test's directory
type FS struct {
	c    *C
	id   int
	mtx  sync.Mutex
	refs int32
	*fs
}

const (
	// DataDir is the name of the directory where test-specific files live;
	// that is, files created specifically for a single test while running the
	// test.
	DataDir = "test_data"

	dirPerms = 0700
)

func newFS(c *C, id int) (*FS, func()) {
	f := &FS{
		c:    c,
		id:   id,
		refs: 0,
	}

	return f.ref()
}

func (f *FS) ref() (*FS, func()) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if f.refs == 0 {
		f.fs = &fs{
			id: f.id,
			c:  f.c,
		}
	}

	f.refs++

	return f, f.unref
}

func (f *FS) unref() {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.refs--
	if f.refs > 0 {
		return
	}

	f.fs.destroy()
	f.fs = nil
}

type fs struct {
	id      int
	c       *C
	setup   sync.Once
	dataDir string
}

func (f *fs) destroy() {
	// Don't remove the data dir if the test failed
	if f.dataDir != "" && !f.c.Failed() {
		f.cleanDir(f.dataDir)
	}
}

// Path gets the absolute path of the test's data dir joined with the parts.
// The parent dirs are created, assuming that the last part is the file.
func (f *fs) Path(parts ...string) string {
	path := f.join(parts...)

	parent := filepath.Join(path, "..")
	err := f.mkdirAll(parent)
	f.c.Must.Nil(err)

	return path
}

func (f *fs) join(parts ...string) string {
	ps := []string{f.GetDataDir()}
	ps = append(ps, parts...)
	return filepath.Join(ps...)
}

// FileExists checks that the given file exists
func (f *fs) FileExists(parts ...string) bool {
	path := f.join(parts...)
	exists, err := cfs.FileExists(path)
	f.c.Must.Nil(err)
	return f.c.True(exists, "%s does not exist", path)
}

// NotFileExists checks that the given file does not exist
func (f *fs) NotFileExists(parts ...string) bool {
	path := f.join(parts...)
	exists, err := cfs.FileExists(path)
	f.c.Must.Nil(err)
	return f.c.False(exists, "%s exists when it shouldn't", path)
}

// DirExists checks that the given directory exists
func (f *fs) DirExists(parts ...string) bool {
	path := f.join(parts...)
	exists, err := cfs.DirExists(path)
	f.c.Must.Nil(err)
	return f.c.True(exists, "%s does not exist", path)
}

// NotDirExists checks that the given directory exists
func (f *fs) NotDirExists(parts ...string) bool {
	path := f.join(parts...)
	exists, err := cfs.DirExists(path)
	f.c.Must.Nil(err)
	return f.c.False(exists, "%s exists when it shouldn't", path)
}

// WriteFile writes the given contents to the given path in the test's data
// dir, creating everything as necessary.
func (f *fs) WriteFile(path string, contents []byte) {
	err := cfs.Write(f.Path(path), contents)
	f.c.Must.Nil(err)
}

// SWriteFile is like WriteFile, but it works on strings
func (f *fs) SWriteFile(path, contents string) {
	f.WriteFile(path, []byte(contents))
}

// ReadFile reads the contents of the file at given path in the test's data
// dir.
func (f *fs) ReadFile(parts ...string) []byte {
	path := f.Path(parts...)

	b, err := ioutil.ReadFile(path)
	f.c.Must.Nil(err)

	return b
}

// SReadFile is like ReadFile, but it works on strings.
func (f *fs) SReadFile(parts ...string) string {
	return string(f.ReadFile(parts...))
}

// ContentsEqual checks that the contents of the given file exactly equal the
// given byte slice.
func (f *fs) ContentsEqual(path string, b []byte) {
	a := f.ReadFile(path)
	f.c.Equal(b, a)
}

// SContentsEqual is like ContentsEqual, but with strings.
func (f *fs) SContentsEqual(path string, b string) {
	a := f.SReadFile(path)
	f.c.Equal(b, a)
}

// GetDataDir gets the test's data directory. On the first call, this also clears
// out any data directory that existed previously, giving your test a clean
// space to run.
//
// If the DataDir is not found, this panics.
func (f *fs) GetDataDir() string {
	f.setup.Do(func() {
		dDir, err := cfs.FindDirInParents(DataDir)
		f.c.Must.Nil(err)

		relPath, err := filepath.Rel(dDir, f.c.path)
		f.c.Must.Nil(err)

		path := filepath.Clean(filepath.Join("/", relPath))

		dataDir := filepath.Join(dDir, path, f.c.Name())
		dataDir = filepath.Clean(dataDir)

		// Since this is only called once per test, clear it on first access
		f.cleanDir(dataDir)
		f.mkdirAll(dataDir)

		// Set on way out
		f.dataDir = dataDir
	})

	return f.dataDir
}

// CleanDataDir wipes out this test's data directory
func (f *fs) CleanDataDir() {
	f.cleanDir(f.GetDataDir())
}

// DumpTree walks the tree at the given path and writes each file to the test
// log
func (f *fs) DumpTree(path string) {
	f.c.Logf("Tree at `%s`:", path)

	rootPath := f.Path(path)
	filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			f.c.Must.Nil(err)

			rel, err := filepath.Rel(rootPath, path)
			f.c.Must.Nil(err)

			f.c.Logf("\t%s", rel)
			return nil
		})
}

// cleanDir tries to be cautious about removing directories. Rather than just
// assuming that you can RemoveAll the data dir (what if there's a bug
// somewhere and it gets a directory way above the data dir? yikes), it checks
// to be sure that it's removing the right stuff.
func (f *fs) cleanDir(dir string) {
	dDir := fmt.Sprintf("%c%s%c",
		filepath.Separator,
		DataDir,
		filepath.Separator)

	if !strings.Contains(dir, dDir) {
		panic(fmt.Errorf("refusing to remove %s: does not contain `%s`",
			dir, dDir))
	}

	os.RemoveAll(dir)

	// Last one out should attempt to remove parent directories up to
	// `DataDir`.
	for strings.Contains(dir, dDir) {
		dir = filepath.Dir(dir)

		if filepath.Base(dir) == DataDir {
			break
		}

		// If the dir isn't empty, bail. Someone else will get it.
		err := os.Remove(dir)
		if err != nil {
			break
		}
	}
}

// mkdirAll helps in the case where a test is cleaning up (removing
// directories) while another is starting. Rather than doing crazy
// synchronization (possibly even cross-process sync), let's just retry.
// MkdirAll is pretty idempotent: the directories should always be created.
// Retrying on error is pretty safe.
func (*fs) mkdirAll(dir string) (err error) {
	// 15 seems like a reasonable amount of retries
	for i := 0; i < 15; i++ {
		err = os.MkdirAll(dir, dirPerms)
		if err == nil {
			break
		}
	}

	return
}
