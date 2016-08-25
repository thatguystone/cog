package eio

import "github.com/iheartradio/cog"

// Blackhole drops everything and returns nil slices
type Blackhole struct{}

func init() {
	RegisterProducer("blackhole",
		func(Args) (Producer, error) {
			return Blackhole{}, nil
		})

	RegisterConsumer("blackhole",
		func(Args) (Consumer, error) {
			return Blackhole{}, nil
		})
}

// Produce implements Producer.Produce
func (Blackhole) Produce([]byte) {}

// Next implements Consumer.Next
func (Blackhole) Next() ([]byte, error) { return nil, nil }

// Errs implements Producer.Errs
func (Blackhole) Errs() <-chan error { return ClosedErrCh }

// Rotate implements Producer.Rotate
func (Blackhole) Rotate() error { return nil }

// Close implements Consumer/Producer.Close
func (Blackhole) Close() cog.Errors { return cog.Errors{} }
