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
	C     <-chan struct{}
	mtx   sync.Mutex
	exits []Exiter
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

		e.mtx.Lock()
		exits := e.exits
		e.exits = nil
		e.mtx.Unlock()

		for i := len(exits); i > 0; i-- {
			exits[i-1].Exit()
		}

		e.Wait()
	})
}

// AddExiter adds an Exiter to the exit list that is called when Exit() is
// called. Exiters are called in the reverse order that they were added.
func (e *GExit) AddExiter(ex Exiter) {
	e.mtx.Lock()
	e.exits = append(e.exits, ex)
	e.mtx.Unlock()
}
