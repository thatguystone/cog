package cync

import (
	"sync"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestSemaphore(t *testing.T) {
	c := check.New(t)
	s := NewSemaphore(2)

	for i := 0; i < s.count; i++ {
		s.Lock()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s.Lock()
		wg.Done()
	}()

	go func() {
		for i := 0; i < s.count; i++ {
			s.Unlock()
		}
	}()

	s.Lock()
	wg.Wait()

	for i := 0; i < s.count*2; i++ {
		s.Unlock()
	}

	c.Equal(0, s.used)
}

func BenchmarkSemaphore(b *testing.B) {
	s := NewSemaphore(2)

	for i := 0; i < b.N; i++ {
		for i := 0; i < s.count; i++ {
			s.Lock()
		}

		for i := 0; i < s.count; i++ {
			s.Unlock()
		}
	}
}
