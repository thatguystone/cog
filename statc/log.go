package statc

import (
	"fmt"

	"github.com/thatguystone/cog/clog"
)

type logStats struct {
	n Name
	l *clog.Log
}

func (ls *logStats) Snapshot(a Adder) {
	ss := ls.l.Stats()

	for _, s := range ss {
		n := ls.n.Append(s.Module)
		fmt.Println(s.Module)

		for lvl, cnt := range s.Counts {
			a.AddInt(
				n.Append(clog.Level(lvl).String()),
				cnt)
		}
	}
}
