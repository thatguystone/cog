package clog

import (
	"sync"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestNewOutputErrors(t *testing.T) {
	c := check.New(t)

	wg := &sync.WaitGroup{}

	_, err := newOutput(
		&OutputConfig{},
		nil,
		wg)
	c.Error(err)

	_, err = newOutput(
		&OutputConfig{
			Fmt: "json",
		},
		nil,
		wg)
	c.Error(err)

	_, err = newOutput(
		&OutputConfig{
			Prod: "mehhhh",
			Fmt:  "json",
		},
		nil,
		wg)
	c.Error(err)

	_, err = newOutput(
		&OutputConfig{
			Prod: "stdout",
			Fmt:  "human",
			Filters: []FilterConfig{
				FilterConfig{
					Which: "noppperpssadf",
				},
			},
		},
		nil,
		wg)
	c.Error(err)
}

func TestErrorHandling(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)
	cfg.Outputs["error"] = &OutputConfig{
		Prod: "test_errors",
		Fmt:  "json",
	}

	cfg.Modules[""].Outputs = []string{"error"}
	cfg.Modules["clog"] = &ModuleConfig{
		Outputs: []string{"test"},
		Level:   Debug,
	}

	l, err := New(cfg)
	c.MustNotError(err)

	l.Get("so").Info("much fun")
	l.Flush()

	test := c.FS.SReadFile("test")
	c.Contains(test, "i only produce errors")
	c.Contains(test, "yeah, that close failed")
}
