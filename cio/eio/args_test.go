package eio

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestArgsApplyTo(t *testing.T) {
	c := check.New(t)

	s := struct {
		B bool
		I int
	}{}

	args := Args{
		"b": true,
		"i": 1234,
	}

	args.ApplyTo(&s)

	c.True(s.B)
	c.Equal(s.I, 1234)
}

func TestArgsNilApplyTo(t *testing.T) {
	check.New(t)
	s := 0

	// Just don't panic
	var args Args
	args.ApplyTo(&s)
}
