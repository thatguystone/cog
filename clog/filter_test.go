package clog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
)

type rejectMsgFilter struct{}
type rejectDataFilter struct{}

func init() {
	RegisterFilter("rejectMsg", func(eio.Args) (Filter, error) {
		return rejectMsgFilter{}, nil
	})

	RegisterFilter("rejectData", func(eio.Args) (Filter, error) {
		return rejectDataFilter{}, nil
	})
}

func (rejectMsgFilter) Accept(e Entry) bool {
	return !strings.Contains(e.Msg, "reject")
}

func (rejectMsgFilter) Exit() {}

func (rejectDataFilter) Accept(e Entry) bool {
	_, ok := e.Data["reject"]
	return !ok
}

func (rejectDataFilter) Exit() {}

func TestFilterErrors(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		RegisterFilter("rejectMsg", nil)
	})

	_, err := newFilters(Debug, []FilterConfig{
		FilterConfig{Which: "blarg"},
	})
	c.Error(err)
}

func TestFilterDump(t *testing.T) {
	c := check.New(t)

	b := &bytes.Buffer{}
	DumpKnownFilters(b)

	c.Contains(b.String(), "reject")
}

func TestFilterReject(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)

	out := cfg.Outputs["test"]
	out.Level = Info
	out.Filters = []FilterConfig{FilterConfig{Which: "rejectData"}}

	mod := cfg.Modules[""]
	mod.Level = Info
	mod.Filters = []FilterConfig{FilterConfig{Which: "rejectMsg"}}

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Info("reject")
	lg.Info("fun")
	lg.WithKV("reject", true).Info("fun-data")

	test := c.FS.SReadFile("test")
	c.Contains(test, "fun")
	c.NotContains(test, "reject")
	c.NotContains(test, "fun-data")
}
