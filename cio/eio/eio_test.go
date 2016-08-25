package eio

import (
	"runtime"
	"testing"
	"time"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/check"
)

type noopProdCons struct {
	closed bool
}

func init() {
	RegisterProducer("test_noop",
		func(Args) (Producer, error) {
			return &noopProdCons{}, nil
		})

	RegisterConsumer("test_noop",
		func(Args) (Consumer, error) {
			return &noopProdCons{}, nil
		})
}

func TestMain(m *testing.M) {
	check.Main(m)
}

func (n *noopProdCons) Produce([]byte)           {}
func (n *noopProdCons) ProduceTo(string, []byte) {}
func (n *noopProdCons) Next() ([]byte, error)    { return nil, nil }
func (n *noopProdCons) Errs() <-chan error       { return ClosedErrCh }
func (n *noopProdCons) Rotate() error            { return nil }
func (n *noopProdCons) Close() cog.Errors        { n.closed = true; return cog.Errors{} }

func TestEIOFinalizers(t *testing.T) {
	c := check.New(t)

	pr, err := NewProducer("test_noop", Args{})
	c.MustNotError(err)

	tpr, err := NewTopicProducer("test_noop", Args{})
	c.MustNotError(err)

	co, err := NewConsumer("test_noop", Args{})
	c.MustNotError(err)

	prp := pr.(*producer).Producer.(*noopProdCons)
	pr = nil

	tprp := tpr.(*topicProducer).TopicProducer.(*noopProdCons)
	tpr = nil

	cop := co.(*consumer).Consumer.(*noopProdCons)
	co = nil

	for i := 0; i < 50; i++ {
		runtime.GC()
		if prp.closed && cop.closed && tprp.closed {
			break
		}
		time.Sleep(time.Millisecond)
	}

	c.True(prp.closed)
	c.True(tprp.closed)
	c.True(cop.closed)
}

func TestEIOErrors(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		RegisterProducer("test", nil)
		RegisterProducer("TEST", nil)
	})

	c.Panics(func() {
		RegisterConsumer("test", nil)
		RegisterConsumer("TEST", nil)
	})

	_, err := NewProducer("iDontExist", nil)
	c.Error(err)

	_, err = NewTopicProducer("stdout", nil)
	c.Error(err)

	_, err = NewConsumer("iDontExist", nil)
	c.Error(err)
}
