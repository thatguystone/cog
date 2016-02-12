package statc

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// A Timer is used to time how long things take to run
type Timer struct {
	mtx sync.Mutex
	timerSnap

	sampPercent int // Percent of samples to save
	samples     int64Slice

	// Cached names
	nStddev, nMean, nMin, nMax, nCount, nP50, nP75, nP90, nP95 string
}

type int64Slice []int64

type timerSnap struct {
	sum, sumOfSq       int64 // For stddev + mean
	stddev             int64
	mean               int64
	min, max           int64
	count              int64
	p50, p75, p90, p95 int64 // 50th, 75th, and 90th percentiles
}

var (
	timerRand  = rand.New(rand.NewSource(time.Now().Unix()))
	timerReset = timerSnap{
		min: math.MaxInt64,
	}
)

// NewTimer creates a new timer that outputs stats prefixed with the given
// name. The names are cached internally to limit allocations.
//
// The timer also reports percentiles of data by sampling. The given
// `sampPercent` control what percent of samples to save for percentile
// calculations (0 - 100).
func NewTimer(name string, sampPercent int) *Timer {
	if sampPercent < 0 {
		sampPercent = 0
	}

	if sampPercent > 100 {
		sampPercent = 100
	}

	return &Timer{
		timerSnap:   timerReset,
		sampPercent: sampPercent,
		nStddev:     fmt.Sprintf("%s.stddev", name),
		nMean:       fmt.Sprintf("%s.mean", name),
		nMin:        fmt.Sprintf("%s.min", name),
		nMax:        fmt.Sprintf("%s.max", name),
		nCount:      fmt.Sprintf("%s.count", name),
		nP50:        fmt.Sprintf("%s.p50", name),
		nP75:        fmt.Sprintf("%s.p75", name),
		nP90:        fmt.Sprintf("%s.p90", name),
		nP95:        fmt.Sprintf("%s.p95", name),
	}
}

// TimeFunc times how long it takes the given function to run
func (t *Timer) TimeFunc(cb func()) {
	start := time.Now()
	cb()
	t.Add(time.Now().Sub(start))
}

// Add adds timing information
func (t *Timer) Add(dd time.Duration) {
	d := int64(dd)
	sq := d * d

	keep := t.sampPercent > int(timerRand.Int31n(100))

	t.mtx.Lock()

	t.count++
	t.sumOfSq += sq
	t.sum += d

	if d > t.max {
		t.max = d
	}

	if d < t.min {
		t.min = d
	}

	if keep {
		t.samples = append(t.samples, d)
	}

	t.mtx.Unlock()
}

// Snapshot implements Snapshotter
func (t *Timer) Snapshot(a Adder) {
	t.snapshot(a, false)
}

func (t *Timer) snapshot(a Adder, ignoreIfEmpty bool) {
	nsamps := make(int64Slice, 0, t.count)

	t.mtx.Lock()

	ts := t.timerSnap
	t.timerSnap = timerReset

	samps := t.samples
	t.samples = nsamps

	t.mtx.Unlock()

	sort.Sort(samps)

	if ts.min == timerReset.min {
		ts.min = 0
	}

	if ts.count == 0 && ignoreIfEmpty {
		return
	}

	if ts.count > 0 {
		ss := ts.sumOfSq - ((ts.sum * ts.sum) / ts.count)
		ts.stddev = int64(math.Sqrt(float64(ss / ts.count)))
		ts.mean = ts.sum / ts.count
	}

	if len(samps) > 0 {
		l := float64(len(samps))

		ts.p50 = samps[int(math.Ceil(l*.50))-1]
		ts.p75 = samps[int(math.Ceil(l*.75))-1]
		ts.p90 = samps[int(math.Ceil(l*.90))-1]
		ts.p95 = samps[int(math.Ceil(l*.95))-1]
	}

	a.AddInt(t.nStddev, ts.stddev)
	a.AddInt(t.nMean, ts.mean)
	a.AddInt(t.nMin, ts.min)
	a.AddInt(t.nMax, ts.max)
	a.AddInt(t.nCount, ts.count)
	a.AddInt(t.nP50, ts.p50)
	a.AddInt(t.nP75, ts.p75)
	a.AddInt(t.nP90, ts.p90)
	a.AddInt(t.nP95, ts.p95)
}

func (p int64Slice) Len() int           { return len(p) }
func (p int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
