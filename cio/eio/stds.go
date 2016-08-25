package eio

import (
	"bytes"
	"os"

	"github.com/iheartradio/cog"
)

// OutProducer writes to either stdout or stderr, depending on if you create a
// producer with name "stdout" or "stderr".
type OutProducer struct {
	out *os.File
}

func init() {
	RegisterProducer("stdout",
		func(args Args) (Producer, error) {
			return &OutProducer{out: os.Stdout}, nil
		})
	RegisterProducer("stderr",
		func(args Args) (Producer, error) {
			return &OutProducer{out: os.Stderr}, nil
		})
}

// Produce implements Producer.Produce
func (p *OutProducer) Produce(b []byte) {
	f := os.Stdout
	if p.out != nil {
		f = p.out
	}

	b = append(bytes.TrimSpace(b), '\n')
	f.Write(b)
}

// Errs implements Producer.Errs
func (p *OutProducer) Errs() <-chan error { return ClosedErrCh }

// Rotate implements Producer.Rotate
func (p *OutProducer) Rotate() error { return nil }

// Close implements Producer.Close
func (p *OutProducer) Close() (es cog.Errors) { return }
