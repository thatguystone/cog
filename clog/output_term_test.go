package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestOutputTermCoverage(t *testing.T) {
	c := check.New(t)

	o, err := newTermOutput(ConfigArgs{}, HumanFormat{})
	c.MustNotError(err)

	o.Write([]byte(""))
	o.Rotate()
	_ = o.String()
}
