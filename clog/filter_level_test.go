package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestFilterLevelBasic(t *testing.T) {
	c := check.New(t)

	cfg := basicTestConfig(c)

	out := cfg.Outputs["test"]
	out.Level = Info
	out.Filters = []ConfigFilter{
		ConfigFilter{
			Which: lvlFilterName,
			Args:  ConfigArgs{"Level": Info},
		},
	}

	l, err := New(cfg)
	c.MustNotError(err)

	lg := l.Get("test")
	lg.Debug("no")
	lg.Info("yes")
	lg.Error("yes")

	test := c.FS.SReadFile("test")
	c.NotContains(test, "level=debug")
	c.Contains(test, "level=info")
	c.Contains(test, "level=error")
}

func TestFilterLevelErrors(t *testing.T) {
	c := check.New(t)

	_, err := newFilters(Info, []ConfigFilter{
		ConfigFilter{
			Which: lvlFilterName,
			Args:  ConfigArgs{"Level": "ponies"},
		},
	})
	c.Error(err)
}
