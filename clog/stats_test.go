package clog

import (
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestStatsBasic(t *testing.T) {
	c := check.New(t)

	l, err := New(basicTestConfig(c))
	c.MustNotError(err)

	lg := l.Get("one")
	lg.Info("msg")
	for i := 0; i < 10; i++ {
		lg.Warn("warn")
	}

	lg = l.Get("two")
	for i := 0; i < 10; i++ {
		lg.Error("warn")
	}

	lg = l.Get("panic")
	c.Panics(func() {
		lg.Panic("no no no")
	})

	m := map[string]int64{}
	for _, s := range l.Stats() {
		for lvl, cnt := range s.Counts {
			name := fmt.Sprintf("%s.%s", s.Module, Level(lvl).String())
			m[name] = cnt
			c.Logf("%s = %d", name, cnt)
		}
	}

	c.Equal(m["one.debug"], 0)
	c.Equal(m["one.info"], 1)
	c.Equal(m["one.warn"], 10)

	c.Equal(m["two.info"], 0)
	c.Equal(m["two.error"], 10)

	c.Equal(m["panic.panic"], 1)
}
