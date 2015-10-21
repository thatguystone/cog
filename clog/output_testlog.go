package clog

import (
	"bytes"
	"errors"

	"github.com/thatguystone/cog/stack"
)

// TestLogOutput writes to testing.Logf()
type TestLogOutput struct {
	Formatter
	l testLog
}

type testLog interface {
	Log(...interface{})
}

func init() {
	RegisterOutputter("testLog",
		func(a ConfigOutputArgs) (Outputter, error) {
			var log testLog

			l, ok := a["log"]
			if ok {
				log, ok = l.(testLog)
			}

			if !ok {
				return nil, errors.New("provided log argument does " +
					"not implement testing.Log()")
			}

			return &TestLogOutput{
				Formatter: HumanFormat{},
				l:         log,
			}, nil
		})
}

func (o *TestLogOutput) Write(b []byte) error {
	o.l.Log(stack.ClearTestCaller() + string(bytes.TrimSpace(b)))
	return nil
}

// Reopen implements Outputter.Reopen
func (*TestLogOutput) Reopen() error {
	return nil
}

func (*TestLogOutput) String() string {
	return "TestLogOutput"
}
