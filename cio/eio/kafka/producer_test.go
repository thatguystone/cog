package kafka

import (
	"testing"

	"github.com/thatguystone/cog/cio/eio"
)

func TestProducerErrors(t *testing.T) {
	c := newTest(t)

	_, err := eio.NewProducer("kafka", nil)
	c.Error(err)

	_, err = eio.NewProducer("kafka", eio.Args{
		"brokers": make(chan struct{}),
	})
	c.Error(err)
}
