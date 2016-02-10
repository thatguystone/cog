package clog

import (
	"fmt"
	"os"

	"github.com/thatguystone/cog/config"
)

// TermOutput writes messages to the terminal on stdout. By default, this uses
// the human formatter.
type TermOutput struct {
	Formatter
	out *os.File

	Args struct {
		// If this should log to stdout instead of stderr
		Stdout bool
	}
}

func init() {
	fcfg := FormatterConfig{Name: "Human"}

	RegisterOutputter("term", fcfg, newTermOutput)
	RegisterOutputter("terminal", fcfg, newTermOutput)
}

func newTermOutput(a config.Args, f Formatter) (o Outputter, err error) {
	to := &TermOutput{
		Formatter: f,
	}

	err = a.ApplyTo(&to.Args)
	if err == nil {
		if to.Args.Stdout {
			to.out = os.Stdout
		} else {
			to.out = os.Stderr
		}

		o = to
	}

	return
}

func (s *TermOutput) Write(b []byte) error {
	b = append(b, '\n')
	_, err := s.out.Write(b)
	return err
}

// Rotate implements Outputter.Rotate
func (s *TermOutput) Rotate() error {
	return nil
}

// Exit implements Outputter.Exit
func (s *TermOutput) Exit() {}

func (s *TermOutput) String() string {
	return fmt.Sprintf("TermOutput{}")
}
