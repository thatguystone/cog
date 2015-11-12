package cog

import "sync"

// Exit is useful for terminating a group of goroutines that run in a
// for{select{}}. Be sure to `Exit.Add(n)` before starting goroutines, and
// `defer Exit.Done()` in the goroutine.
type Exit struct {
	*GExit
	c    chan struct{}
	once sync.Once
}

// GExit (short for "goroutine exit") is what should be passed to things that
// need to know when to exit but that should not be able to trigger an exit.
type GExit struct {
	sync.WaitGroup
	C <-chan struct{}
}

// Exiter is anything that can cleanup after itself at any arbitrary point in
// time.
type Exiter interface {
	Exit()
}

// NewExit creates a new Exit, useful for ensuring termination of goroutines on
// exit.
func NewExit() *Exit {
	e := &Exit{
		GExit: &GExit{},
		c:     make(chan struct{}),
	}

	e.GExit.C = e.c

	return e
}

// Exit closes C and waits for all goroutines to exit.
func (e *Exit) Exit() {
	e.once.Do(func() {
		close(e.c)
		e.Wait()
	})
}
