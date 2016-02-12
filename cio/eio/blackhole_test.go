package eio

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/config"
)

func TestBlackholeCoverage(t *testing.T) {
	c := check.New(t)

	pr, err := NewProducer("Blackhole", config.Args{})
	c.MustNotError(err)
	pr.Produce(nil)
	pr.Errs()
	pr.Rotate()
	pr.Close()

	co, err := NewConsumer("Blackhole", config.Args{})
	c.MustNotError(err)
	co.Next()
	co.Close()
}
