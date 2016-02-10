package kafka

import (
	"testing"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/check/chlog"
)

func TestProducerMockBasic(t *testing.T) {
	c, log := chlog.New(t)

	exit := cog.NewExit()
	defer exit.Exit()

	p, err := NewProducer(ProducerOpts{
		Mock: true,
		Log:  log.Get("producer"),
		Exit: exit.GExit,
	})
	c.MustNotError(err)

	payload := []byte("message")
	p.Bytes("test", payload)

	pend := p.(*MockProducer).Pending
	msg := <-pend
	c.Equal(msg.Topic, "test")
	c.Equal(msg.Msg, payload)

	exit.Exit()

	// Should not block
	for i := 0; i < cap(pend)*2; i++ {
		p.Bytes("test", payload)
	}
}
