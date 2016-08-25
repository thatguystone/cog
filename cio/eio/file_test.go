package eio

import (
	"os"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestFileBasic(t *testing.T) {
	c := check.New(t)

	p, err := NewProducer("file", Args{
		"Path": c.FS.Path("file"),
	})
	c.MustNotError(err)
	defer p.Close()

	p.Produce([]byte("     test     \n"))
	c.Equal(c.FS.SReadFile("file"), "test\n")

	err = os.Rename(c.FS.Path("file"), c.FS.Path("file.1"))
	c.MustNotError(err)

	err = p.Rotate()
	c.MustNotError(err)

	p.Produce([]byte("     after rotate     \n"))
	c.Equal(c.FS.SReadFile("file"), "after rotate\n")
}

func TestFileOpenErrors(t *testing.T) {
	c := check.New(t)

	c.FS.SWriteFile("file", "")

	path := c.FS.Path("file")
	err := os.Chmod(path, 0)
	c.MustNotError(err)

	_, err = NewProducer("file", Args{
		"Path": path,
	})
	c.MustError(err)
}

func TestFileProduceErrors(t *testing.T) {
	c := check.New(t)

	path := c.FS.Path("file")
	p, err := regdPs["file"](Args{
		"Path": path,
	})
	c.MustNotError(err)

	p.(*FileProducer).f.Close()
	p.Produce([]byte("message"))

	select {
	case err = <-p.Errs():
		c.Error(err)
	default:
		c.Fatal("no error sent")
	}

	c.Equal(c.FS.SReadFile("file"), "")
}

func TestFileRotateErrors(t *testing.T) {
	c := check.New(t)

	path := c.FS.Path("file")
	p, err := NewProducer("file", Args{
		"Path": path,
	})
	c.MustNotError(err)
	defer p.Close()

	p.Produce([]byte("     before rotate     \n"))

	err = os.Rename(c.FS.Path("file"), c.FS.Path("file.1"))
	c.MustNotError(err)

	c.FS.SWriteFile("file", "")
	err = os.Chmod(path, 0)
	c.MustNotError(err)

	err = p.Rotate()
	c.Error(err)

	p.Produce([]byte("     after rotate     \n"))
	c.Equal(c.FS.SReadFile("file.1"), "before rotate\nafter rotate\n")
}
