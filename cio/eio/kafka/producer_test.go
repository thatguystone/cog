package kafka

import (
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
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

func TestProducerCloseErrors(t *testing.T) {
	c := newTest(t)

	pr, err := newProducer(eio.Args{
		"brokers": []string{kafkaAddr},
		"topic":   check.GetTestName(),
	})
	c.MustNotError(err)

	kp := pr.(*Producer)

	kp.errs <- fmt.Errorf("here's an error")
	c.Error(<-pr.Errs())

	kp.errs <- fmt.Errorf("close error")
	es := pr.Close()
	c.Error(es.Error())
}
