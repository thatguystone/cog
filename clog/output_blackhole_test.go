package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestOutputBlackholeCoverage(t *testing.T) {
	c := check.New(t)
	o := BlackholeOutput{}

	_, err := o.FormatEntry(Entry{})
	c.MustNotError(err)

	o.Write([]byte(""))
	o.Rotate()
	_ = o.String()
}
