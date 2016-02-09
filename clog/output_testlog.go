package clog

import (
	"bytes"
	"errors"

	"github.com/thatguystone/cog/stack"
)

// TestLogOutput writes to testing.Logf()
//
// This output is for capturing the output of your application to the test log,
// so that if a test fails, you have the entire application log handy, otherwise
// it's all hidden.
//
// This is a special output in that it must be configured programmatically. You
// configure it directly in Config itself, as follows:
//
//     Config{
//         Outputs: map[string]*OutputConfig{
//             "testlog": {
//                 Which: "TestLog",
//                 Level: clog.Debug,
//                 Args: ConfigArgs{
//                     "log": t, // Anything with a Log(...interface{}) method
//                 },
//             },
//         },
//         Modules: map[string]*ModuleConfig{
//             "": {
//                 Outputs: []string{"testlog"},
//                 Level:   clog.Debug,
//             },
//         },
//     }
//
// In the above example, you pass a testing.TB as the argument; really, it will
// accept anything with a `Log(...interface{})` method. All log output will be
// directed to this function as a single string.
//
// Or, equivalently:
//
//     chlog.New(t)
type TestLogOutput struct {
	Formatter
	l testLog
}

type testLog interface {
	Log(...interface{})
}

func init() {
	RegisterOutputter(
		"TestLog",
		FormatterConfig{Name: "Human"},
		func(a ConfigArgs, f Formatter) (Outputter, error) {
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
				Formatter: f,
				l:         log,
			}, nil
		})
}

func (o *TestLogOutput) Write(b []byte) error {
	o.l.Log(stack.ClearTestCaller() + string(bytes.TrimSpace(b)))
	return nil
}

// Rotate implements Outputter.Rotate
func (*TestLogOutput) Rotate() error {
	return nil
}

// Exit implements Outputter.Exit
func (*TestLogOutput) Exit() {}

func (*TestLogOutput) String() string {
	return "TestLogOutput"
}
