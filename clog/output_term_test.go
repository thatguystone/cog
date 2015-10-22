package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestOutputTermCoverage(t *testing.T) {
	c := check.New(t)

	o, err := newTermOutput(ConfigOutputArgs{})
	c.MustNotError(err)

	o.Write([]byte(""))
	o.Reopen()
	_ = o.String()
}
