package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/config"
)

func TestOutputTermCoverage(t *testing.T) {
	c := check.New(t)

	o, err := newTermOutput(config.Args{}, HumanFormat{})
	c.MustNotError(err)

	o.Write([]byte(""))
	o.Rotate()
	_ = o.String()
}
