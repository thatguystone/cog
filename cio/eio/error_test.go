package eio

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestErrorBasic(t *testing.T) {
	c := check.New(t)

	p, err := NewProducer("test_errors", nil)
	c.MustNotError(err)
	defer p.Close()

	p.Produce(nil)
	c.Error(p.Rotate())
	c.Error(<-p.Errs())

	es := p.Close()
	c.Error(es.Error())
}
