package kafka

import (
	"testing"

	"github.com/thatguystone/cog/cio/eio"
)

func TestConsumerErrors(t *testing.T) {
	c := newTest(t)

	_, err := eio.NewConsumer("kafka", nil)
	c.Error(err)

	_, err = eio.NewConsumer("kafka", eio.Args{
		"brokers": make(chan struct{}),
	})
	c.Error(err)
}

func TestConsumerInvalidTopic(t *testing.T) {
	c := newTest(t)

	_, err := eio.NewConsumer("kafka", eio.Args{
		"brokers": []string{kafkaAddr},
		"topic":   "this topic does not exist",
	})
	c.Error(err)
}
