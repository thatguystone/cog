package cync

import "sync"

// Semaphore implements a basic counting semaphore
type Semaphore struct {
	used  int
	count int
	mtx   sync.Mutex
	cond  *sync.Cond
}

// NewSemaphore creates a new counting semaphore
func NewSemaphore(count int) *Semaphore {
	s := &Semaphore{
		count: count,
	}

	s.cond = sync.NewCond(&s.mtx)

	return s
}

// Lock implements sync.Locker.Lock()
func (s *Semaphore) Lock() {
	s.mtx.Lock()

	for s.used == s.count {
		s.cond.Wait()
	}

	s.used++

	s.mtx.Unlock()
}

// Unlock implements sync.Locker.Unlock()
func (s *Semaphore) Unlock() {
	s.mtx.Lock()

	s.used--
	if s.used < 0 {
		s.used = 0
	}

	s.cond.Signal()

	s.mtx.Unlock()
}
