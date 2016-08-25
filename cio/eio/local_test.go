package eio

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestLocals(t *testing.T) {
	c := check.New(t)
	args := Args{
		"topic": check.GetTestName(),
	}

	pr, err := NewProducer("local", args)
	c.MustNotError(err)

	co, err := NewConsumer("local", args)
	c.MustNotError(err)

	pr.Produce([]byte("test"))

	b, err := co.Next()
	c.MustNotError(err)

	c.Equal(string(b), "test")

	es := pr.Close()
	c.MustNotError(es.Error())

	es = co.Close()
	c.MustNotError(es.Error())

	_, ok := localTopics[check.GetTestName()]
	c.False(ok)

	// Should not block
	b, err = co.Next()
	c.MustNotError(err)
	c.Len(b, 0)

	pr.Produce([]byte("test"))

	c.Equal(nil, <-pr.Errs())
	c.NotError(pr.Rotate())
}

func TestLocalTopicProducer(t *testing.T) {
	const topic = "some-random-topic"

	c := check.New(t)
	tp, err := NewTopicProducer("local", nil)
	c.MustNotError(err)
	defer tp.Close()

	co, err := NewConsumer("local", Args{
		"topic": topic,
	})
	c.MustNotError(err)

	tp.ProduceTo("no-one-is-listening", []byte("nothing"))

	tp.ProduceTo(topic, []byte("super important info"))
	msg, err := co.Next()
	c.Equal(string(msg), "super important info")
}

func TestLocalErrors(t *testing.T) {
	c := check.New(t)
	args := Args{
		"topic": make(chan []byte),
	}

	_, err := NewProducer("local", args)
	c.Error(err)

	_, err = NewTopicProducer("local", args)
	c.Error(err)

	_, err = NewConsumer("local", args)
	c.Error(err)
}

func TestLocalMessagesOnClose(t *testing.T) {
	c := check.New(t)
	args := Args{
		"topic": check.GetTestName(),
	}

	pr, err := NewProducer("local", args)
	c.MustNotError(err)

	co, err := NewConsumer("local", args)
	c.MustNotError(err)

	pr.Produce([]byte("test"))

	es := pr.Close()
	c.MustNotError(es.Error())

	es = co.Close()
	c.Error(es.Error())
}
