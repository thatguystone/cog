package eio

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestTestLogBasic(t *testing.T) {
	c := check.New(t)

	p, err := NewProducer("testlog", Args{
		"log": c,
	})
	c.MustNotError(err)

	p.Produce([]byte("test"))

	c.Equal(nil, <-p.Errs())
	c.NotError(p.Rotate())
	errs := p.Close()
	c.NotError(errs.Error())
}

func TestTestLogErrors(t *testing.T) {
	c := check.New(t)

	_, err := NewProducer("testlog", nil)
	c.Error(err)

	_, err = NewProducer("testlog", Args{
		"log": "hooray",
	})
	c.Error(err)
}
