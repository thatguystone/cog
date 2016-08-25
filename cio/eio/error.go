package eio

import (
	"fmt"
	"sync"

	"github.com/iheartradio/cog"
)

// An ErrorProducer only ever returns errors. It's useful for testing error
// handling.
type ErrorProducer struct {
	ch   chan error
	once sync.Once
}

func init() {
	RegisterProducer("test_errors",
		func(args Args) (Producer, error) {
			p := &ErrorProducer{
				ch: make(chan error, 1),
			}

			return p, nil
		})
}

// Produce implements Producer.Produce
func (p *ErrorProducer) Produce([]byte) {
	p.ch <- fmt.Errorf("i only produce errors")
}

// Rotate implements Producer.Rotate
func (*ErrorProducer) Rotate() error {
	return fmt.Errorf("i refuse to be told to rotate things")
}

// Errs implements Producer.Errs
func (p *ErrorProducer) Errs() <-chan error {
	return p.ch
}

// Close implements Producer.Close
func (p *ErrorProducer) Close() (errs cog.Errors) {
	p.once.Do(func() {
		close(p.ch)
		errs.Add(fmt.Errorf("yeah, that close failed"))
	})
	return
}
