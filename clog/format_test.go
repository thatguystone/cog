package clog

import (
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
)

func TestRegisterFormatterErrors(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		RegisterFormatter("JSON",
			func(args eio.Args) (Formatter, error) {
				return nil, nil
			})
	})

	_, err := newFormatter("lulz what", eio.Args{})
	c.MustError(err)
}
