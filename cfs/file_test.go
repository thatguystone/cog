package cfs_test

import (
	"testing"

	"github.com/iheartradio/cog/cfs"
	"github.com/iheartradio/cog/check"
)

func TestCreate(t *testing.T) {
	c := check.New(t)

	path := c.FS.Path("really/long/path/with/parents/file")
	f, err := cfs.Create(path)
	c.MustNotError(err)
	f.Close()

	exists, err := cfs.FileExists(path)
	c.MustNotError(err)
	c.True(exists)
}

func TestCreateError(t *testing.T) {
	c := check.New(t)

	_, err := cfs.Create("/nope/not/allowed")
	c.Error(err)
}

func TestWrite(t *testing.T) {
	c := check.New(t)

	err := cfs.Write(c.FS.Path("file"), []byte("some stuff!"))
	c.MustNotError(err)

	c.FS.SContentsEqual("file", "some stuff!")
}

func TestWriteError(t *testing.T) {
	c := check.New(t)

	err := cfs.Write("/this/is/not/allowed", []byte("some stuff!"))
	c.Error(err)
}

func TestCopy(t *testing.T) {
	c := check.New(t)

	c.FS.SWriteFile("file", "crazy contents")

	err := cfs.Copy(c.FS.Path("file"), c.FS.Path("copy"))
	c.MustNotError(err)

	c.FS.SContentsEqual("copy", "crazy contents")
}

func TestCopyErrors(t *testing.T) {
	c := check.New(t)

	c.FS.SWriteFile("file", "crazy contents")

	err := cfs.Copy("/i/do/not/exist", c.FS.Path("copy"))
	c.Error(err)

	err = cfs.Copy(c.FS.Path("file"), "/this/is/not/allowed")
	c.Error(err)

	c.FS.SWriteFile("dest", "crazy contents")
}
