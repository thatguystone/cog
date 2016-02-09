// Package stats implements runtime process stats and status reporting.
//
// It provides the basics (timers, gauges, counters), stats sinks, interfaces
// for fetching current stats, and so on.
package stats

import (
	"fmt"
	"sort"
	"sync"

	"github.com/thatguystone/cog/clog"
)

// S is a stats aggregator.
type S struct {
	cfg Config
	log *clog.Logger

	mtx      sync.Mutex
	snappers []snapshotter
}

// NewS creates a new stats aggregator
func NewS(cfg Config, log *clog.Logger) *S {
	return &S{
		cfg: cfg,
		log: log,
	}
}

// AddSnapshotter binds a snapshotter to this S.
func (s *S) AddSnapshotter(name string, snapper Snapshotter) {
	l := len(s.snappers)
	i := sort.Search(l, func(i int) bool {
		return s.snappers[i].name >= name
	})

	if i < l && s.snappers[i].name == name {
		panic(fmt.Errorf("a Snapshotter with name `%s` already exists", name))
	}

	s.snappers = append(s.snappers, snapshotter{})

	copy(s.snappers[i+1:], s.snappers[i:])
	s.snappers[i] = snapshotter{
		name:    name,
		snapper: snapper,
	}
}

// NewTimer creates a timer that's bound to this S
func (s *S) NewTimer(name string, sampPercent int) *Timer {
	t := NewTimer(name, sampPercent)
	s.AddSnapshotter(name, t)
	return t
}

// NewCounter creates a counter that's bound to this S
func (s *S) NewCounter(name string, reset bool) *Counter {
	c := NewCounter(reset)
	s.AddSnapshotter(name, c)
	return c
}

// NewGauge creates a gauge that's bound to this S
func (s *S) NewGauge(name string) *Gauge {
	g := new(Gauge)
	s.AddSnapshotter(name, g)
	return g
}

// NewBoolGauge creates a bool gauge that's bound to this S
func (s *S) NewBoolGauge(name string) *BoolGauge {
	g := new(BoolGauge)
	s.AddSnapshotter(name, g)
	return g
}

// NewFloatGauge creates a float gauge that's bound to this S
func (s *S) NewFloatGauge(name string) *FloatGauge {
	g := new(FloatGauge)
	s.AddSnapshotter(name, g)
	return g
}

// NewStringGauge creates a string gauge that's bound to this S
func (s *S) NewStringGauge(name string) *StringGauge {
	g := new(StringGauge)
	s.AddSnapshotter(name, g)
	return g
}

func (s *S) snapshot() (snap Snapshot) {
	for _, sn := range s.snappers {
		snap.Take(sn.name, sn.snapper)
	}

	return
}
