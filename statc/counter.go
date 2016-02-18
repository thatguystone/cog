package statc

import "sync/atomic"

// A Counter is used to count things
type Counter struct {
	v     int64
	reset bool
}

// NewCounter creates a new, unbound counter
func NewCounter(reset bool) *Counter {
	return &Counter{
		reset: reset,
	}
}

// Add adds the given value to the current counter
func (c *Counter) Add(i int64) {
	atomic.AddInt64(&c.v, i)
}

// Inc is the equivalent of Add(1)
func (c *Counter) Inc() {
	c.Add(1)
}

// Dec is the equivalent of Add(-1)
func (c *Counter) Dec() {
	c.Add(-1)
}

// Get gets the current value of the counter
func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.v)
}

// Snapshot implements Snapshotter
func (c *Counter) Snapshot(a Adder) {
	var old int64
	if c.reset {
		old = atomic.SwapInt64(&c.v, 0)
	} else {
		old = c.Get()
	}

	a.AddInt(Name{}, old)
}
