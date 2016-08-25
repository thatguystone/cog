package eio

import (
	"os"
	"testing"

	"github.com/iheartradio/cog/check"
)

func testOut(t *testing.T, name string) {
	c := check.New(t)

	p, err := regdPs[name](Args{})
	c.MustNotError(err)
	defer p.Close()

	path := c.FS.Path("file")
	f, err := os.Create(path)
	c.MustNotError(err)
	defer f.Close()

	p.(*OutProducer).out = f

	p.Produce([]byte("   test   \n"))
	c.Equal(c.FS.SReadFile("file"), "test\n")

	c.Equal(nil, <-p.Errs())
	c.NotError(p.Rotate())
	errs := p.Close()
	c.NotError(errs.Error())
}

func TestStdoutBasic(t *testing.T) {
	testOut(t, "stdout")
}

func TestStderrBasic(t *testing.T) {
	testOut(t, "stderr")
}
