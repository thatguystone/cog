package statc

import (
	"testing"
	"time"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/check/chlog"
	"github.com/iheartradio/cog/cio/eio"
	"github.com/iheartradio/cog/clog"
	"github.com/iheartradio/cog/ctime"
)

type sTest struct {
	*S
	log  *clog.Ctx
	exit *cog.Exit
}

func TestMain(m *testing.M) {
	check.Main(m)
}

func newTest(t *testing.T, cfg *Config) (*check.C, *sTest) {
	c, log := chlog.New(t)
	st := &sTest{
		log:  log,
		exit: cog.NewExit(),
	}

	if cfg == nil {
		cfg = &Config{
			SnapshotInterval:  ctime.Millisecond,
			HTTPSamplePercent: 100,
			StatusKey:         statusKey,
			Outputs: []OutputConfig{
				OutputConfig{
					Prod: "file",
					ProdArgs: eio.Args{
						"path": c.FS.Path("stats"),
					},
					Fmt: "json",
				},
			},
		}
	}

	var err error
	st.S, err = NewS(*cfg, log.Get("statc"), st.exit.GExit)
	c.MustNotError(err)

	return c, st
}

func TestStatsBasic(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	tr := st.NewTimer("test.timer", 50)
	tre := st.NewTimer("test.timer_empty", 50)
	cs := st.NewCounter("test.counter.save", false)
	cr := st.NewCounter("test.counter.reset", true)
	g := st.NewGauge("test.gauge")
	bg := st.NewBoolGauge("test.gauge_bool")
	fg := st.NewFloatGauge("test.gauge_float")
	sg := st.NewStringGauge("test.gauge_string")

	for i := 0; i < 100; i++ {
		tr.Add(time.Millisecond)
		cs.Inc()
		cr.Inc()
	}

	for i := 0; i < 50; i++ {
		cs.Dec()
		cr.Dec()
	}

	c.Equal(cs.Get(), 50)
	c.Equal(cr.Get(), 50)

	g.Set(123)
	c.Equal(g.Get(), 123)
	g.Set(345)

	bg.Set(true)
	c.Equal(bg.Get(), true)
	bg.Set(false)
	c.Equal(bg.Get(), false)
	bg.Set(true)

	fg.Set(1.234)
	c.Equal(fg.Get(), 1.234)
	fg.Set(4.567)

	sg.Set("abcd")
	c.Equal(sg.Get(), "abcd")
	sg.Set("efgh")

	snap := st.snapshot()
	c.True(len(snap) > 0, "len(snap) = %d", len(snap))

	c.Equal(snap.Get(tr.nMean).Val.(int64), time.Millisecond)
	c.Equal(snap.Get(tr.nMin).Val.(int64), time.Millisecond)
	c.Equal(snap.Get(tr.nMax).Val.(int64), time.Millisecond)
	c.Equal(snap.Get(tr.nCount).Val.(int64), 100)
	c.Equal(snap.Get(tr.nStddev).Val.(int64), 0)
	c.Equal(snap.Get(tr.nP50).Val.(int64), time.Millisecond)
	c.Equal(snap.Get(tr.nP75).Val.(int64), time.Millisecond)
	c.Equal(snap.Get(tr.nP90).Val.(int64), time.Millisecond)
	c.Equal(snap.Get(tr.nP95).Val.(int64), time.Millisecond)

	c.Equal(snap.Get(tre.nMin).Val.(int64), 0)
	c.Equal(snap.Get(tre.nMax).Val.(int64), 0)

	c.Equal(snap.Get(st.Names("test", "counter", "save")).Val.(int64), 50)
	c.Equal(snap.Get(st.Names("test", "counter", "reset")).Val.(int64), 50)
	c.Equal(cs.Get(), 50)
	c.Equal(cr.Get(), 0)

	c.Equal(snap.Get(st.Names("test", "gauge")).Val.(int64), 345)
	c.Equal(g.Get(), 345)

	c.Equal(snap.Get(st.Names("test", "gauge_bool")).Val.(bool), true)
	c.Equal(bg.Get(), true)

	c.Equal(snap.Get(st.Names("test", "gauge_float")).Val.(float64), 4.567)
	c.Equal(fg.Get(), 4.567)

	c.Equal(snap.Get(st.Names("test", "gauge_string")).Val.(string), "efgh")
	c.Equal(sg.Get(), "efgh")
}

func TestStatsPrefixed(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	st.NewCounter("top-level", true)

	pst := st.Prefixed("long.prefix...")
	pst.NewCounter("sub", true)

	pst = st.Prefixed("...another..prefix...")
	pst.NewCounter("magic", true)

	exists := func(n Name) bool {
		for _, s := range st.snappers {
			if s.name.s == n.s {
				return true
			}
		}

		return false
	}

	c.True(exists(st.Name("another.prefix.magic")))
	c.True(exists(st.Name("long.prefix.sub")))
	c.True(exists(st.Name("top-level")))
}

func TestStatsAlreadyExists(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	c.Panics(func() {
		st.NewTimer("test.timer", 50)
		st.NewTimer("test.timer", 50)
	})
}

func TestStatsErrors(t *testing.T) {
	c, clog := chlog.New(t)
	exit := cog.NewExit()
	cfg := Config{
		Outputs: []OutputConfig{
			OutputConfig{
				Prod: "iDontExist",
			},
		},
	}

	_, err := NewS(cfg, clog.Get("stats"), exit.GExit)
	c.Error(err)
}

func TestStatsGet(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	snap := st.snapshot()
	c.Equal(snap.Get(st.Names("i don't exist")).Val, nil)
}

func TestStatsSnapshotting(t *testing.T) {
	c, st := newTest(t, &Config{
		SnapshotInterval: ctime.Millisecond,
	})
	defer st.exit.Exit()

	st.NewTimer("timer", 10)

	snap := st.Snapshot()
	c.Until(time.Second, func() bool {
		snap = st.Snapshot()
		return len(snap) > 0
	})

	c.Until(time.Second, func() bool {
		return &snap[0] != &st.Snapshot()[0]
	})
}

func TestStatsNameErrors(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	c.Panics(func() {
		st.Names("", "", "")
	})

	c.Panics(func() {
		st.Name("")
	})
}
