package kafka

import (
	"testing"
	"time"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/check/chlog"
	"github.com/thatguystone/cog/wrap/kafka/local"
)

func TestProducerBasic(t *testing.T) {
	c, log := chlog.New(t)

	exit := cog.NewExit()
	defer exit.Exit()

	h := local.Run()
	defer h.Close()

	p, err := NewProducer(ProducerOpts{
		Brokers: []string{h.Addr()},
		Log:     log.Get("producer"),
		Exit:    exit.GExit,
	})
	c.MustNotError(err)

	p.Bytes(check.GetTestName(), []byte("message!"))

	exit.Exit()
	c.Until(time.Second, func() bool {
		return p.(*producer).cl.Closed()
	})

	// Should not panic or anything
	p.Bytes(check.GetTestName(), []byte("message!"))
}

func TestProducerErrors(t *testing.T) {
	c, log := chlog.New(t)

	exit := cog.NewExit()
	defer exit.Exit()

	_, err := NewProducer(ProducerOpts{
		Brokers: []string{"herlp derp"},
		Log:     log.Get("producer"),
		Exit:    exit.GExit,
	})
	c.Error(err)
}
