package clog

import (
	"fmt"
	"os"
)

// TermOutput writes messages to the terminal on stdout
type TermOutput struct {
	Formatter
}

func init() {
	RegisterOutputter("term", newTermOutput)
	RegisterOutputter("terminal", newTermOutput)
}

func newTermOutput(a ConfigOutputArgs) (o Outputter, err error) {
	hf := HumanFormat{}

	err = a.ApplyTo(&hf.Args)
	if err == nil {
		o = &TermOutput{
			Formatter: hf,
		}
	}

	return
}

func (s *TermOutput) Write(b []byte) error {
	_, err := os.Stderr.Write(b)
	return err
}

// Reopen implements Outputter.Reopen
func (s *TermOutput) Reopen() error {
	return nil
}

func (s *TermOutput) String() string {
	return fmt.Sprintf("TermOutput{}")
}
