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

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/clog"
)

// S is a stats aggregator.
type S struct {
	*s
	prefix string
}

type s struct {
	cfg  Config
	log  *clog.Log
	exit *cog.GExit

	outExit *cog.Exit
	outs    []*output

	mtx      sync.Mutex
	snappers []snapshotter
	lastSnap Snapshot
}

// NewS creates a new stats aggregator
func NewS(cfg Config, log *clog.Log, exit *cog.GExit) (*S, error) {
	cfg.setDefaults()

	s := &S{
		s: &s{
			cfg:  cfg,
			log:  log,
			exit: exit,

			// Nest exits so that, if there's an error setting up any output,
			// all the outputs can be terminated by killing this
			outExit: cog.NewExit(),
			outs:    make([]*output, 0, len(cfg.Outputs)),
		},
	}

	for _, cfg := range cfg.Outputs {
		var out *output
		out, err := newOutput(cfg, log, s.outExit.GExit)
		if err != nil {
			s.outExit.Exit()
			return nil, fmt.Errorf("failed to create output %s: %v", cfg.Prod, err)
		}

		s.outs = append(s.outs, out)
	}

	s.watchRuntime()
	s.exit.Add(1)
	go s.run()

	return s, nil
}

// Prefixed returns a new S that prefixes all stats with the given prefix
func (s *S) Prefixed(prefix string) *S {
	sc := *s
	sc.prefix = JoinNoEscape(s.prefix, prefix)
	return &sc
}

// Name returns a new name, appended to this S's prefix. The given name is not
// escaped.
func (s *S) Name(name string) Name {
	if name == "" {
		panic(fmt.Errorf("cannot create a name with an empty string"))
	}

	return newName(s.prefix).Append(name)
}

// Names returns a new Name with the given names joined with this S's prefix.
// The names are escaped individually.
func (s *S) Names(names ...string) Name {
	return s.Name(JoinPath("", names...))
}

// AddSnapshotter binds a snapshotter to this S.
func (s *S) AddSnapshotter(name Name, snapper Snapshotter) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	l := len(s.snappers)
	i := sort.Search(l, func(i int) bool {
		return s.snappers[i].name.Str() >= name.Str()
	})

	if i < l && s.snappers[i].name == name {
		panic(fmt.Errorf("a Snapshotter with name `%s` already exists", name.Str()))
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
	n := s.Name(name)
	t := NewTimer(n, sampPercent)
	s.AddSnapshotter(n, t)
	return t
}

// NewCounter creates a counter that's bound to this S
func (s *S) NewCounter(name string, reset bool) *Counter {
	c := NewCounter(reset)
	s.AddSnapshotter(s.Name(name), c)
	return c
}

// NewGauge creates a gauge that's bound to this S
func (s *S) NewGauge(name string) *Gauge {
	g := new(Gauge)
	s.AddSnapshotter(s.Name(name), g)
	return g
}

// NewBoolGauge creates a bool gauge that's bound to this S
func (s *S) NewBoolGauge(name string) *BoolGauge {
	g := new(BoolGauge)
	s.AddSnapshotter(s.Name(name), g)
	return g
}

// NewFloatGauge creates a float gauge that's bound to this S
func (s *S) NewFloatGauge(name string) *FloatGauge {
	g := new(FloatGauge)
	s.AddSnapshotter(s.Name(name), g)
	return g
}

// NewStringGauge creates a string gauge that's bound to this S
func (s *S) NewStringGauge(name string) *StringGauge {
	g := new(StringGauge)
	s.AddSnapshotter(s.Name(name), g)
	return g
}

// AddLog adds log stats at the given path
func (s *S) AddLog(name string, l *clog.Ctx) {
	n := s.Name(name)
	s.AddSnapshotter(n, &logStats{n: n, l: l})
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
	snap = make(Snapshot, 0, cap(s.lastSnap))

	for _, sn := range s.snappers {
		snap.Take(sn.name, sn.snapper)
	}

	return
}
