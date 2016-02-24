package statc

import (
	"math"
	"math/big"
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
	nStddev, nMean, nMin, nMax, nCount, nP50, nP75, nP90, nP95 Name
}

type int64Slice []int64

type timerSnap struct {
	sum                int64
	sumOfSq            *big.Int
	stddev             int64
	min, max           int64
	count              int64
	p50, p75, p90, p95 int64 // 50th, 75th, 90th, and 95th percentiles
}

var (
	// This amounts to ~2x speedup
	bigPool = sync.Pool{
		New: func() interface{} {
			return big.NewInt(0)
		},
	}

	// Rand is not safe to use concurrently, and the global rand functions all
	// use a mutex. Another pool to the rescue!
	randPool = sync.Pool{
		New: func() interface{} {
			return rand.New(rand.NewSource(time.Now().UnixNano()))
		},
	}
)

func bigInt(i int64) *big.Int {
	return bigPool.Get().(*big.Int).SetInt64(i)
}

// NewTimer creates a new timer that outputs stats prefixed with the given
// name. The names are cached internally to limit allocations.
//
// The timer also reports percentiles of data by sampling. The given
// `sampPercent` controls what percent of samples to save for percentile
// calculations (0 - 100).
func NewTimer(name Name, sampPercent int) (t *Timer) {
	if sampPercent < 0 {
		sampPercent = 0
	}

	if sampPercent > 100 {
		sampPercent = 100
	}

	t = &Timer{
		sampPercent: sampPercent,
		nStddev:     name.Join("stddev"),
		nMean:       name.Join("mean"),
		nMin:        name.Join("min"),
		nMax:        name.Join("max"),
		nCount:      name.Join("count"),
		nP50:        name.Join("p50"),
		nP75:        name.Join("p75"),
		nP90:        name.Join("p90"),
		nP95:        name.Join("p95"),
	}

	t.timerSnap.reset()

	return t
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
	sq := bigInt(d)
	sq.Mul(sq, sq)

	rand := randPool.Get().(*rand.Rand)
	keep := t.sampPercent > int(rand.Int31n(100))
	randPool.Put(rand)

	t.mtx.Lock()

	t.count++
	t.sumOfSq.Add(t.sumOfSq, sq)
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

	bigPool.Put(sq)
}

// Snapshot implements Snapshotter
func (t *Timer) Snapshot(a Adder) {
	t.snapshot(a, false)
}

func (t *Timer) snapshot(a Adder, ignoreIfEmpty bool) {
	nsamps := make(int64Slice, 0, t.count)

	tsr := timerSnap{}
	tsr.reset()

	t.mtx.Lock()

	ts := t.timerSnap
	t.timerSnap = tsr

	samps := t.samples
	t.samples = nsamps

	t.mtx.Unlock()

	sort.Sort(samps)

	if ts.min == math.MaxInt64 {
		ts.min = 0
	}

	if ts.count == 0 && ignoreIfEmpty {
		return
	}

	var mean int64
	if ts.count > 0 {
		// This is the equivalent of: ts.sumOfSq - ((ts.sum * ts.sum) / ts.count)
		sumSq := bigInt(ts.sum)
		sumSq.Mul(sumSq, sumSq)

		cnt := bigInt(ts.count)
		sumSq.Div(sumSq, cnt)
		bigPool.Put(cnt)

		ss := ts.sumOfSq.Sub(ts.sumOfSq, sumSq).Int64()
		bigPool.Put(sumSq)

		ts.stddev = int64(math.Sqrt(float64(ss / ts.count)))
		mean = ts.sum / ts.count
	}

	if len(samps) > 0 {
		l := float64(len(samps))

		ts.p50 = samps[int(math.Ceil(l*.50))-1]
		ts.p75 = samps[int(math.Ceil(l*.75))-1]
		ts.p90 = samps[int(math.Ceil(l*.90))-1]
		ts.p95 = samps[int(math.Ceil(l*.95))-1]
	}

	a.AddInt(t.nStddev, ts.stddev)
	a.AddInt(t.nMean, mean)
	a.AddInt(t.nMin, ts.min)
	a.AddInt(t.nMax, ts.max)
	a.AddInt(t.nCount, ts.count)
	a.AddInt(t.nP50, ts.p50)
	a.AddInt(t.nP75, ts.p75)
	a.AddInt(t.nP90, ts.p90)
	a.AddInt(t.nP95, ts.p95)
	ts.free()
}

func (p int64Slice) Len() int           { return len(p) }
func (p int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (ts *timerSnap) reset() {
	ts.sumOfSq = bigInt(0)
	ts.min = math.MaxInt64
}

func (ts *timerSnap) free() {
	bigPool.Put(ts.sumOfSq)
}
