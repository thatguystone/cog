package cfs_test

import (
	"testing"

	"github.com/thatguystone/cog/cfs"
	"github.com/thatguystone/cog/check"
)

func TestCreate(t *testing.T) {
	c := check.New(t)

	fs, clean := c.FS()
	defer clean()

	path := fs.Path("really/long/path/with/parents/file")
	f, err := cfs.Create(path)
	c.Must.Nil(err)
	f.Close()

	exists, err := cfs.FileExists(path)
	c.Must.Nil(err)
	c.True(exists)
}

func TestCreateError(t *testing.T) {
	c := check.New(t)

	_, err := cfs.Create("/nope/not/allowed")
	c.NotNil(err)
}

func TestWrite(t *testing.T) {
	c := check.New(t)

	fs, clean := c.FS()
	defer clean()

	err := cfs.Write(fs.Path("file"), []byte("some stuff!"))
	c.Must.Nil(err)

	fs.SContentsEqual("file", "some stuff!")
}

func TestWriteError(t *testing.T) {
	c := check.New(t)

	err := cfs.Write("/this/is/not/allowed", []byte("some stuff!"))
	c.NotNil(err)
}

func TestCopy(t *testing.T) {
	c := check.New(t)

	fs, clean := c.FS()
	defer clean()

	fs.SWriteFile("file", "crazy contents")

	err := cfs.Copy(fs.Path("file"), fs.Path("copy"))
	c.Must.Nil(err)

	fs.SContentsEqual("copy", "crazy contents")
}

func TestCopyErrors(t *testing.T) {
	c := check.New(t)

	fs, clean := c.FS()
	defer clean()

	fs.SWriteFile("file", "crazy contents")

	err := cfs.Copy("/i/do/not/exist", fs.Path("copy"))
	c.NotNil(err)

	err = cfs.Copy(fs.Path("file"), "/this/is/not/allowed")
	c.NotNil(err)

	fs.SWriteFile("dest", "crazy contents")
}
