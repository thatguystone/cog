package kafka

import (
	"math"
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
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

func TestConsumerErrorsCoverage(t *testing.T) {
	c := newTest(t)

	co, err := newConsumer(eio.Args{
		"brokers": []string{kafkaAddr},
		"topic":   check.GetTestName(),
	})
	c.MustNotError(err)

	kc := co.(*Consumer)
	kc.exit.Add(1)
	kc.consume(math.MaxInt32)

	_, err = co.Next()
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
