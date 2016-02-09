package clog

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

type errorOutput struct {
	check.Errorer
	f Outputter
}

type exitOutput struct {
	exited bool
}

var errOut errorOutput

func init() {
	fcfg := FormatterConfig{
		Name: "Human",
		Args: ConfigArgs{
			"ShortTime": true,
		},
	}

	RegisterOutputter("errOut",
		fcfg,
		func(a ConfigArgs, fmttr Formatter) (Outputter, error) {
			if errOut.Fail() {
				return nil, errors.New("nope, not gonna happen")
			}

			f, err := newFileOutputter(a, fmttr)
			if err != nil {
				return nil, err
			}

			return &errorOutput{f: f}, nil
		})

	RegisterOutputter("exitOut",
		fcfg,
		func(a ConfigArgs, f Formatter) (Outputter, error) {
			return &exitOutput{}, nil
		})
}

func (o *errorOutput) FormatEntry(e Entry) ([]byte, error) {
	if o.Fail() {
		return nil, errors.New("i don't want to format that")
	}

	return o.f.FormatEntry(e)
}

func (o *errorOutput) Write(b []byte) error {
	if o.Fail() {
		return errors.New("i don't want to write that")
	}

	return o.f.Write(b)
}

func (o *errorOutput) Rotate() error {
	if o.Fail() {
		return errors.New("i don't want to rotate that")
	}

	return o.f.Rotate()
}

func (o *errorOutput) Exit() {}

func (o *errorOutput) String() string {
	return "ErrorOutput"
}

func (o *exitOutput) FormatEntry(e Entry) ([]byte, error) { return nil, nil }
func (o *exitOutput) Write(b []byte) error                { return nil }
func (o *exitOutput) Rotate() error                       { return nil }
func (o *exitOutput) Exit()                               { o.exited = true }
func (o *exitOutput) String() string                      { return "exitOutput" }

func TestOutputDump(t *testing.T) {
	c := check.New(t)

	b := &bytes.Buffer{}
	DumpKnownOutputs(b)

	c.Contains(b.String(), "file")
}

func TestOutputErrors(t *testing.T) {
	c := check.New(t)

	c.Panic(func() {
		RegisterOutputter("errOut", FormatterConfig{}, nil)
	})

	_, err := newOutput(&OutputConfig{
		Which: "file",
		Formatter: FormatterConfig{
			Name: "nope",
		},
	})
	c.Error(err)

	cfg := basicTestConfig(c)

	out := cfg.Outputs["test"]
	out.Which = "errOut"
	out.Level = Info

	mod := cfg.Modules[""]
	mod.Level = Info

	var l *Log
	for l == nil {
		l, _ = New(cfg)
	}

	lg := l.Get("test")
	for !strings.Contains(c.FS.SReadFile("test"), "test") {
		lg.Info("test")
	}
}
