package eio

import (
	"runtime"
	"testing"
	"time"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/config"
)

type noopProdCons struct {
	closed bool
}

func init() {
	RegisterProducer("test_noop",
		func(config.Args) (Producer, error) {
			return &noopProdCons{}, nil
		})

	RegisterConsumer("test_noop",
		func(config.Args) (Consumer, error) {
			return &noopProdCons{}, nil
		})
}

func TestMain(m *testing.M) {
	check.Main(m)
}

func (n *noopProdCons) Produce([]byte)        {}
func (n *noopProdCons) Next() ([]byte, error) { return nil, nil }
func (n *noopProdCons) Errs() <-chan error    { return nil }
func (n *noopProdCons) Rotate() error         { return nil }
func (n *noopProdCons) Close() cog.Errors     { n.closed = true; return cog.Errors{} }

func TestEIOFinalizers(t *testing.T) {
	c := check.New(t)

	pr, err := NewProducer("test_noop", config.Args{})
	c.MustNotError(err)

	co, err := NewConsumer("test_noop", config.Args{})
	c.MustNotError(err)

	prp := pr.(*producer).Producer.(*noopProdCons)
	pr = nil

	cop := co.(*consumer).Consumer.(*noopProdCons)
	co = nil

	for i := 0; i < 50; i++ {
		runtime.GC()
		if prp.closed && cop.closed {
			break
		}
		time.Sleep(time.Millisecond)
	}

	c.True(prp.closed)
	c.True(cop.closed)
}

func TestEIOErrors(t *testing.T) {
	c := check.New(t)

	c.Panic(func() {
		RegisterProducer("test", nil)
		RegisterProducer("TEST", nil)
	})

	c.Panic(func() {
		RegisterConsumer("test", nil)
		RegisterConsumer("TEST", nil)
	})

	_, err := NewProducer("iDontExist", nil)
	c.Error(err)

	_, err = NewConsumer("iDontExist", nil)
	c.Error(err)
}