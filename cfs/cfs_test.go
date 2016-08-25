package cfs_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/iheartradio/cog/cfs"
	"github.com/iheartradio/cog/check"
)

func TestMain(m *testing.M) {
	check.Main(m)
}

func TestFindInParents(t *testing.T) {
	c := check.New(t)

	_, err := cfs.FindInParents("idontexist")
	c.Error(err)

	_, err = cfs.FindInParents("README.md")
	c.NotError(err)
}

func TestFindDirInParents(t *testing.T) {
	c := check.New(t)

	_, err := cfs.FindDirInParents("idontexist")
	c.Error(err)

	_, err = cfs.FindDirInParents("cog")
	c.NotError(err)
}

func TestFileExists(t *testing.T) {
	c := check.New(t)

	fs, err := filepath.Glob("*")
	c.MustNotError(err)
	c.MustTrue(len(fs) > 0)

	ex, err := cfs.FileExists(fs[0])
	c.MustNotError(err)
	c.True(ex, "%s does not exist", fs[0])

	ex, err = cfs.FileExists("/i/dont/exist")
	c.MustNotError(err)
	c.False(ex)
}

func TestDirExists(t *testing.T) {
	c := check.New(t)

	dir, err := os.Getwd()
	c.MustNotError(err)

	ex, err := cfs.DirExists(dir)
	c.MustNotError(err)
	c.True(ex, "%s does not exist", dir)

	ex, err = cfs.DirExists("/i/dont/exist/")
	c.MustNotError(err)
	c.False(ex)
}

func TestTempFile(t *testing.T) {
	c := check.New(t)

	f, err := cfs.TempFile("tmp-", "aac")
	c.MustNotError(err)

	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	c.Contains(f.Name(), "tmp-")
	c.Contains(f.Name(), ".aac")
}

func TestImportPath(t *testing.T) {
	c := check.New(t)

	_, err := cfs.ImportPath("does not exist", false)
	c.MustError(err)

	_, filename, _, _ := runtime.Caller(0)
	path, err := cfs.ImportPath(filename, false)
	c.MustNotError(err, "filename=%s", filename)
	c.Equal(path, "github.com/iheartradio/cog/cfs")
}
