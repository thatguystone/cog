package eio

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/config"
)

func TestLocals(t *testing.T) {
	c := check.New(t)
	args := config.Args{
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

	c.Len(localTopics, 0)

	// Should not block
	b, err = co.Next()
	c.MustNotError(err)
	c.Len(b, 0)

	pr.Produce([]byte("test"))

	c.Equal(nil, pr.Errs())
	c.NotError(pr.Rotate())
}

func TestLocalErrors(t *testing.T) {
	c := check.New(t)
	args := config.Args{
		"topic": make(chan []byte),
	}

	_, err := NewProducer("local", args)
	c.Error(err)

	_, err = NewConsumer("local", args)
	c.Error(err)
}

func TestLocalMessagesOnClose(t *testing.T) {
	c := check.New(t)
	args := config.Args{
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
