package stats

import (
	"testing"
	"time"

	"github.com/thatguystone/cog/check/chlog"
)

func TestStatsBasic(t *testing.T) {
	c, clog := chlog.New(t)
	s := NewS(Config{}, clog.Get("stats"))

	tr := s.NewTimer("test.timer", 50)
	s.NewTimer("test.timer_empty", 50)
	cs := s.NewCounter("test.counter.save", false)
	cr := s.NewCounter("test.counter.reset", true)
	g := s.NewGauge("test.gauge")
	bg := s.NewBoolGauge("test.gauge_bool")
	fg := s.NewFloatGauge("test.gauge_float")
	sg := s.NewStringGauge("test.gauge_string")

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

	snap := s.snapshot()
	c.True(len(snap) > 0, "len(snap) = %d", len(snap))

	c.Equal(snap.Get("test.timer.mean").Val.(int64), time.Millisecond)
	c.Equal(snap.Get("test.timer.min").Val.(int64), time.Millisecond)
	c.Equal(snap.Get("test.timer.max").Val.(int64), time.Millisecond)
	c.Equal(snap.Get("test.timer.count").Val.(int64), 100)
	c.Equal(snap.Get("test.timer.stddev").Val.(int64), 0)
	c.Equal(snap.Get("test.timer.p50").Val.(int64), time.Millisecond)
	c.Equal(snap.Get("test.timer.p75").Val.(int64), time.Millisecond)
	c.Equal(snap.Get("test.timer.p90").Val.(int64), time.Millisecond)
	c.Equal(snap.Get("test.timer.p95").Val.(int64), time.Millisecond)

	c.Equal(snap.Get("test.timer_empty.min").Val.(int64), 0)
	c.Equal(snap.Get("test.timer_empty.max").Val.(int64), 0)

	c.Equal(snap.Get("test.counter.save").Val.(int64), 50)
	c.Equal(snap.Get("test.counter.reset").Val.(int64), 50)
	c.Equal(cs.Get(), 50)
	c.Equal(cr.Get(), 0)

	c.Equal(snap.Get("test.gauge").Val.(int64), 345)
	c.Equal(g.Get(), 345)

	c.Equal(snap.Get("test.gauge_bool").Val.(bool), true)
	c.Equal(bg.Get(), true)

	c.Equal(snap.Get("test.gauge_float").Val.(float64), 4.567)
	c.Equal(fg.Get(), 4.567)

	c.Equal(snap.Get("test.gauge_string").Val.(string), "efgh")
	c.Equal(sg.Get(), "efgh")
}

func TestStatsAlreadyExists(t *testing.T) {
	c, clog := chlog.New(t)
	s := NewS(Config{}, clog.Get("stats"))

	c.Panic(func() {
		s.NewTimer("test.timer", 50)
		s.NewTimer("test.timer", 50)
	})
}

func TestStatsGet(t *testing.T) {
	c, clog := chlog.New(t)
	s := NewS(Config{}, clog.Get("stats"))

	snap := s.snapshot()
	c.Equal(snap.Get("i don't exist").Val, nil)
}
