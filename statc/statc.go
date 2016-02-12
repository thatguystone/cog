// Package statc implements runtime process stats and status reporting.
//
// It provides the basics (timers, gauges, counters), stats sinks, interfaces
// for fetching current stats, and so on.
package statc

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/clog"
)

// S is a stats aggregator.
type S struct {
	cfg  Config
	log  *clog.Logger
	exit *cog.GExit

	outExit *cog.Exit
	outs    []*output

	mtx      sync.Mutex
	snappers []snapshotter
	lastSnap Snapshot
}

// NewS creates a new stats aggregator
func NewS(cfg Config, log *clog.Logger, exit *cog.GExit) (s *S, err error) {
	cfg.setDefaults()

	s = &S{
		cfg:  cfg,
		log:  log,
		exit: exit,

		// Nest exits so that, if there's an error setting up any output, all
		// the outputs can be terminated by killing this
		outExit: cog.NewExit(),
		outs:    make([]*output, 0, len(cfg.Outputs)),
	}

	for _, cfg := range cfg.Outputs {
		var out *output
		out, err = newOutput(cfg, log, s.outExit.GExit)
		if err != nil {
			s.outExit.Exit()
			s = nil

			err = fmt.Errorf("failed to create output %s: %v", cfg.Prod, err)
			return
		}

		s.outs = append(s.outs, out)
	}

	s.exit.Add(1)
	go s.run()

	return
}

// AddSnapshotter binds a snapshotter to this S.
func (s *S) AddSnapshotter(name string, snapper Snapshotter) {
	name = CleanPath(name)
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

// Snapshot gets the last snapshot. If len(snap) == 0, then no snapshot has
// been taken yet.
func (s *S) Snapshot() (snap Snapshot) {
	s.mtx.Lock()
	snap = s.lastSnap
	s.mtx.Unlock()
	return
}

func (s *S) run() {
	defer func() {
		s.outExit.Exit()
		s.exit.Done()
	}()

	t := time.NewTicker(s.cfg.SnapshotInterval.D())
	defer t.Stop()

	for {
		select {
		case <-t.C:
			s.doSnapshot()

		case <-s.exit.C:
			return
		}
	}
}

func (s *S) doSnapshot() {
	snap := s.snapshot()
	s.mtx.Lock()
	s.lastSnap = snap
	s.mtx.Unlock()

	for _, out := range s.outs {
		out.send(snap)
	}
}

func (s *S) snapshot() (snap Snapshot) {
	snap = make(Snapshot, 0, len(s.lastSnap))

	for _, sn := range s.snappers {
		snap.Take(sn.name, sn.snapper)
	}

	return
}
