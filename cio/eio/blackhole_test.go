package eio

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestBlackholeCoverage(t *testing.T) {
	c := check.New(t)

	pr, err := NewProducer("Blackhole", Args{})
	c.MustNotError(err)
	pr.Produce(nil)
	pr.Errs()
	pr.Rotate()
	pr.Close()

	co, err := NewConsumer("Blackhole", Args{})
	c.MustNotError(err)
	co.Next()
	co.Close()
}
