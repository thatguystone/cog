package statc

import (
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

func TestTimerValidate(t *testing.T) {
	c := check.New(t)

	tr := NewTimer(Name{s: "timer"}, 100000)
	c.Equal(tr.sampPercent, 100)

	tr = NewTimer(Name{s: "timer"}, -1)
	c.Equal(tr.sampPercent, 0)
}

func TestTimerTimeFunc(t *testing.T) {
	c := check.New(t)

	tr := NewTimer(Name{s: "timer"}, 100)
	tr.TimeFunc(func() {
		time.Sleep(time.Microsecond)
	})

	c.Len(tr.samples, 1)
}

func TestTimerOverflow(t *testing.T) {
	c := check.New(t)

	name := newName("timer")
	tr := NewTimer(name, 100)

	for i := 0; i < 50000; i++ {
		tr.Add(30295988630 + time.Duration(i))
	}

	snap := make(Snapshot, 0, 100)
	snap.Take(name, tr)
	c.True(len(snap) > 0, "len(snap) = %d", len(snap))

	stddev := snap.Get(tr.nStddev).Val.(int64)
	c.Equal(stddev, 14433)
}

func BenchmarkTimerAdd(b *testing.B) {
	b.ReportAllocs()

	name := newName("timer")
	tr := NewTimer(name, 100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tr.Add(30295988630 + time.Duration(i))
	}
}
