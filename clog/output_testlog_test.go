package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestTestLogBasic(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)

	out := cfg.Outputs["test"]
	out.Which = "testLog"
	out.Args["log"] = t

	l, err := New(cfg)
	c.MustNotError(err)

	l.Get("").Infod(Data{"fun": 1}, "TEST!")
}

func TestTestLogErrors(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)

	out := cfg.Outputs["test"]

	out.Which = "testLog"
	_, err := New(cfg)
	c.Error(err)

	out.Args["log"] = cfg
	_, err = New(cfg)
	c.Error(err)
}

func TestTestLogCoverage(t *testing.T) {
	check.New(t)

	o := TestLogOutput{}
	o.Reopen()
	_ = o.String()
}
