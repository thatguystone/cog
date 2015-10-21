package cog

import "sync"

// Exit is useful for terminating a group of goroutines that run in a
// for{select{}}. Be sure to `Exit.Add(n)` before starting goroutines, and
// `defer Exit.Done()` in the goroutine.
type Exit struct {
	sync.WaitGroup
	C    <-chan struct{}
	c    chan struct{}
	once sync.Once
}

// NewExit creates a new Exit, useful for ensuring termination of goroutines on
// exit.
func NewExit() *Exit {
	e := &Exit{
		c: make(chan struct{}),
	}

	e.C = e.c

	return e
}

// Exit closes C and waits for all goroutines to exit.
func (e *Exit) Exit() {
	e.once.Do(func() {
		close(e.c)
		e.Wait()
	})
}
