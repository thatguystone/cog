package statc

import (
	"sync"
	"sync/atomic"
)

// A Gauge is a single value that can be set at will
type Gauge struct {
	v int64
}

// A BoolGauge is a single bool value that can be set at will
type BoolGauge struct {
	mtx sync.Mutex
	v   bool
}

// A FloatGauge is a single float value that can be set at will
type FloatGauge struct {
	mtx sync.Mutex
	v   float64
}

// A StringGauge is a single string value that can be set at will
type StringGauge struct {
	mtx sync.Mutex
	v   string
}

// Set sets the current value of this gauge
func (g *Gauge) Set(v int64) {
	atomic.StoreInt64(&g.v, v)
}

// Get gets the current value of the gauge
func (g *Gauge) Get() int64 {
	return atomic.LoadInt64(&g.v)
}

// Snapshot implements Snapshotter
func (g *Gauge) Snapshot(a Adder) {
	a.AddInt("", g.Get())
}

// Set sets the current value of this gauge
func (g *BoolGauge) Set(v bool) {
	g.mtx.Lock()
	g.v = v
	g.mtx.Unlock()
}

// Get gets the current value of the gauge
func (g *BoolGauge) Get() (v bool) {
	g.mtx.Lock()
	v = g.v
	g.mtx.Unlock()
	return
}

// Snapshot implements Snapshotter
func (g *BoolGauge) Snapshot(a Adder) {
	a.AddBool("", g.Get())
}

// Set sets the current value of this gauge
func (g *FloatGauge) Set(v float64) {
	g.mtx.Lock()
	g.v = v
	g.mtx.Unlock()
}

// Get gets the current value of the gauge
func (g *FloatGauge) Get() (v float64) {
	g.mtx.Lock()
	v = g.v
	g.mtx.Unlock()
	return
}

// Snapshot implements Snapshotter
func (g *FloatGauge) Snapshot(a Adder) {
	a.AddFloat("", g.Get())
}

// Set sets the current value of this gauge
func (g *StringGauge) Set(v string) {
	g.mtx.Lock()
	g.v = v
	g.mtx.Unlock()
}

// Get gets the current value of the gauge
func (g *StringGauge) Get() (v string) {
	g.mtx.Lock()
	v = g.v
	g.mtx.Unlock()
	return
}

// Snapshot implements Snapshotter
func (g *StringGauge) Snapshot(a Adder) {
	a.AddString("", g.Get())
}
