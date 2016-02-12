package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/config"
)

func TestRegisterFormatterErrors(t *testing.T) {
	c := check.New(t)

	c.Panic(func() {
		RegisterFormatter("JSON",
			func(args config.Args) (Formatter, error) {
				return nil, nil
			})
	})

	_, err := newFormatter("lulz what", config.Args{})
	c.MustError(err)
}
