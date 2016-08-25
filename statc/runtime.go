package statc

import (
	"runtime"
	"sync"
	"time"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/ctime"
)

type cgoCalls struct{}
type numGoroutines struct{}

type memStats struct {
	intv ctime.HumanDuration
	exit *cog.GExit

	mtx sync.Mutex
	ms  runtime.MemStats

	// Cached names
	nAlloc, nTotalAlloc, nSys, nLookups, nMallocs, nFrees                             Name
	nHeapAlloc, nHeapSys, nHeapIdle, nHeapInuse, nHeapReleased, nHeapObjects          Name
	nStackInuse, nStackSys                                                            Name
	nMSpanInuse, nMSpanSys, nMCacheInuse, nMCacheSys, nBuckHashSys, nGCSys, nOtherSys Name
	nNextGC, nLastGC, nPauseTotalNs, nLastPause, nNumGC, nGCCPUFraction               Name
}

func (s *S) watchRuntime() {
	if !s.cfg.disableRuntimeStats {
		s.AddSnapshotter(s.Names("runtime", "num_cgo_calls"), cgoCalls{})
		s.AddSnapshotter(s.Names("runtime", "num_goroutines"), numGoroutines{})
	}

	if s.cfg.MemStatsInterval > 0 {
		n := s.Names("runtime", "mem")
		s.AddSnapshotter(n, newMemStats(n, s.cfg.MemStatsInterval, s.exit))
	}
}

func newMemStats(n Name, intv ctime.HumanDuration, exit *cog.GExit) *memStats {
	ms := &memStats{
		exit: exit,
		intv: intv,

		nAlloc:      n.Join("general", "alloc"),
		nTotalAlloc: n.Join("general", "total"),
		nSys:        n.Join("general", "sys"),
		nLookups:    n.Join("general", "lookups"),
		nMallocs:    n.Join("general", "mallocs"),
		nFrees:      n.Join("general", "frees"),

		nHeapAlloc:    n.Join("heap", "alloc"),
		nHeapSys:      n.Join("heap", "sys"),
		nHeapIdle:     n.Join("heap", "idle"),
		nHeapInuse:    n.Join("heap", "inuse"),
		nHeapReleased: n.Join("heap", "released"),
		nHeapObjects:  n.Join("heap", "objs"),

		nStackInuse: n.Join("stack", "inuse"),
		nStackSys:   n.Join("stack", "sys"),

		nMSpanInuse:  n.Join("low", "mspan", "inuse"),
		nMSpanSys:    n.Join("low", "mspan", "sys"),
		nMCacheInuse: n.Join("low", "mcache", "inuse"),
		nMCacheSys:   n.Join("low", "mcache", "sys"),
		nBuckHashSys: n.Join("low", "buckhashsys"),
		nGCSys:       n.Join("low", "gcsys"),
		nOtherSys:    n.Join("low", "othersys"),

		nNextGC:        n.Join("gc", "next"),
		nLastGC:        n.Join("gc", "last_ts"),
		nLastPause:     n.Join("gc", "last_pause"),
		nPauseTotalNs:  n.Join("gc", "pause_total"),
		nNumGC:         n.Join("gc", "count"),
		nGCCPUFraction: n.Join("gc", "cpu_perc"),
	}

	ms.exit.Add(1)
	go ms.run()

	return ms
}

func (ms *memStats) run() {
	defer ms.exit.Done()

	t := time.NewTicker(ms.intv.D())
	defer t.Stop()

	for {
		select {
		case <-t.C:
			ms.mtx.Lock()
			runtime.ReadMemStats(&ms.ms)
			ms.mtx.Unlock()

		case <-ms.exit.C:
			return
		}
	}
}

func (ms *memStats) Snapshot(a Adder) {
	ms.mtx.Lock()
	ms.mtx.Unlock()

	a.AddInt(ms.nAlloc, int64(ms.ms.Alloc))
	a.AddInt(ms.nTotalAlloc, int64(ms.ms.TotalAlloc))
	a.AddInt(ms.nSys, int64(ms.ms.Sys))
	a.AddInt(ms.nLookups, int64(ms.ms.Lookups))
	a.AddInt(ms.nMallocs, int64(ms.ms.Mallocs))
	a.AddInt(ms.nFrees, int64(ms.ms.Frees))

	a.AddInt(ms.nHeapAlloc, int64(ms.ms.HeapAlloc))
	a.AddInt(ms.nHeapSys, int64(ms.ms.HeapSys))
	a.AddInt(ms.nHeapIdle, int64(ms.ms.HeapIdle))
	a.AddInt(ms.nHeapInuse, int64(ms.ms.HeapInuse))
	a.AddInt(ms.nHeapReleased, int64(ms.ms.HeapReleased))
	a.AddInt(ms.nHeapObjects, int64(ms.ms.HeapObjects))

	a.AddInt(ms.nStackInuse, int64(ms.ms.StackInuse))
	a.AddInt(ms.nStackSys, int64(ms.ms.StackSys))

	a.AddInt(ms.nMSpanInuse, int64(ms.ms.MSpanInuse))
	a.AddInt(ms.nMSpanSys, int64(ms.ms.MSpanSys))
	a.AddInt(ms.nMCacheInuse, int64(ms.ms.MCacheInuse))
	a.AddInt(ms.nMCacheSys, int64(ms.ms.MCacheSys))
	a.AddInt(ms.nBuckHashSys, int64(ms.ms.BuckHashSys))
	a.AddInt(ms.nGCSys, int64(ms.ms.GCSys))
	a.AddInt(ms.nOtherSys, int64(ms.ms.OtherSys))

	a.AddInt(ms.nNextGC, int64(ms.ms.NextGC))
	a.AddInt(ms.nLastGC, int64(ms.ms.LastGC))
	a.AddInt(ms.nLastPause, int64(ms.ms.PauseNs[(ms.ms.NumGC+255)%256]))
	a.AddInt(ms.nPauseTotalNs, int64(ms.ms.PauseTotalNs))
	a.AddInt(ms.nNumGC, int64(ms.ms.NumGC))
	a.AddFloat(ms.nGCCPUFraction, ms.ms.GCCPUFraction)
}

func (cgoCalls) Snapshot(a Adder) {
	a.AddInt(Name{}, runtime.NumCgoCall())
}

func (numGoroutines) Snapshot(a Adder) {
	a.AddInt(Name{}, int64(runtime.NumGoroutine()))
}
