package clog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thatguystone/cog/check"
)

type rejectMsgFilter struct{}
type rejectDataFilter struct{}

func init() {
	RegisterFilter("rejectMsg", rejectMsgFilter{})
	RegisterFilter("rejectData", rejectDataFilter{})
}

func (rejectMsgFilter) Accept(e Entry) bool {
	return !strings.Contains(e.Msg, "reject")
}

func (rejectDataFilter) Accept(e Entry) bool {
	_, ok := e.Data["reject"]
	return !ok
}

func TestFilterErrors(t *testing.T) {
	c := check.New(t)

	c.Panic(func() {
		RegisterFilter("rejectMsg", nil)
	})

	_, err := newFilters(Debug, []string{"blarg"})
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
	out.Filters = []string{"rejectData"}

	mod := cfg.Modules[""]
	mod.Level = Info
	mod.Filters = []string{"rejectMsg"}

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Info("reject")
	lg.Info("fun")
	lg.Infod(Data{"reject": true}, "fun-data")

	test := c.FS.SReadFile("test")
	c.Contains(test, "fun")
	c.NotContains(test, "reject")
	c.NotContains(test, "fun-data")
}
