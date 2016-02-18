package clog

import "sync/atomic"

// Stats about which module has logged how many times at which levels
type Stats struct {
	Module string
	Counts [Fatal]int64 // Fatal obviously can't make it here, so use it as len
}

func (s *Stats) flush() (sc Stats) {
	sc.Module = s.Module

	for i := range s.Counts {
		sc.Counts[i] = atomic.SwapInt64(&s.Counts[i], 0)
	}

	return
}
