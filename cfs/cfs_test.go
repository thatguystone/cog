package cfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thatguystone/cog/cfs"
	"github.com/thatguystone/cog/check"
)

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
