package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestLogfmtCoverage(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")

	lg.Infod(Data{"key with spaces": 1}, "")

	conts := c.FS.SReadFile("test")
	c.Contains(conts, "failed to write log entry")
	c.Contains(conts, "host=")
}
