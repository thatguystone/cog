package statc

import "github.com/iheartradio/cog/clog"

type logStats struct {
	n Name
	l *clog.Ctx
}

func (ls *logStats) Snapshot(a Adder) {
	ss := ls.l.Stats()

	for _, s := range ss {
		n := ls.n.Append(s.Module)

		for lvl, cnt := range s.Counts {
			a.AddInt(
				n.Append(clog.Level(lvl).String()),
				cnt)
		}
	}
}
