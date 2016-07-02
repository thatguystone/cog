package cfs_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/thatguystone/cog/cfs"
	"github.com/thatguystone/cog/check"
)

func TestFindInParents(t *testing.T) {
	c := check.New(t)

	_, err := cfs.FindInParents("idontexist")
	c.NotNil(err)

	_, err = cfs.FindInParents("README.md")
	c.Nil(err)
}

func TestFindDirInParents(t *testing.T) {
	c := check.New(t)

	_, err := cfs.FindDirInParents("idontexist")
	c.NotNil(err)

	_, err = cfs.FindDirInParents("cog")
	c.Nil(err)
}

func TestFileExists(t *testing.T) {
	c := check.New(t)

	fs, err := filepath.Glob("*")
	c.Must.Nil(err)
	c.Must.True(len(fs) > 0)

	ex, err := cfs.FileExists(fs[0])
	c.Must.Nil(err)
	c.True(ex, "%s does not exist", fs[0])

	ex, err = cfs.FileExists("/i/dont/exist")
	c.Must.Nil(err)
	c.False(ex)
}

func TestDirExists(t *testing.T) {
	c := check.New(t)

	dir, err := os.Getwd()
	c.Must.Nil(err)

	ex, err := cfs.DirExists(dir)
	c.Must.Nil(err)
	c.True(ex, "%s does not exist", dir)

	ex, err = cfs.DirExists("/i/dont/exist/")
	c.Must.Nil(err)
	c.False(ex)
}

func TestTempFile(t *testing.T) {
	c := check.New(t)

	f, err := cfs.TempFile("tmp-", "aac")
	c.Must.Nil(err)

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
	c.Must.NotNil(err)

	_, filename, _, _ := runtime.Caller(0)
	path, err := cfs.ImportPath(filename, false)
	c.Must.Nil(err, "filename=%s", filename)
	c.Equal(path, "github.com/thatguystone/cog/cfs")
}
