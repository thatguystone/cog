package clog

import (
	"fmt"
	"os"
)

// TermOutput writes messages to the terminal on stdout
type TermOutput struct {
	Formatter
	out *os.File

	args struct {
		// If this should log to stdout instead of stderr
		Stdout bool
	}
}

func init() {
	RegisterOutputter("term", newTermOutput)
	RegisterOutputter("terminal", newTermOutput)
}

func newTermOutput(a ConfigArgs) (o Outputter, err error) {
	hf := HumanFormat{}

	err = a.ApplyTo(&hf.Args)
	if err == nil {
		to := &TermOutput{
			Formatter: hf,
		}

		err = a.ApplyTo(&to.args)
		if err == nil {
			if to.args.Stdout {
				to.out = os.Stdout
			} else {
				to.out = os.Stderr
			}

			o = to
		}
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
