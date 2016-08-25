package eio

import (
	"bytes"
	"errors"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/stack"
)

// TestLogProducer writes each to the test log. This producer MUST receive a
// single argument, "log", which includes the test's log (typically the t in
// `t *testing.T`).
type TestLogProducer struct {
	l testLogger
}

type testLogger interface {
	Log(...interface{})
}

func init() {
	RegisterProducer(
		"TestLog",
		func(args Args) (Producer, error) {
			var log testLogger

			l, ok := args["log"]
			if ok {
				log, ok = l.(testLogger)
			}

			if !ok {
				return nil, errors.New("provided log argument does " +
					"not implement testing.Log()")
			}

			return &TestLogProducer{
				l: log,
			}, nil
		})
}

// Produce implements Producer.Produce
func (p *TestLogProducer) Produce(b []byte) {
	p.l.Log(stack.ClearTestCaller() + string(bytes.TrimSpace(b)))
}

// Errs implements Producer.Errs
func (*TestLogProducer) Errs() <-chan error { return ClosedErrCh }

// Rotate implements Producer.Rotate
func (*TestLogProducer) Rotate() error { return nil }

// Close implements Producer.Close
func (*TestLogProducer) Close() (errs cog.Errors) { return }
