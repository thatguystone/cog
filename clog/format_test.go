package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/cio/eio"
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
