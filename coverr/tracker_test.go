package coverr

import (
	"testing"
)

func TestTrackerBasic(t *testing.T) {
	trk := new(Tracker)

	EventuallyNil(t, 5, func(i int) error {
		return trk.Err()
	})
}

func TestErrCoverage(t *testing.T) {
	_ = Err.Error()
}

func BenchmarkTracker(b *testing.B) {
	var tr Tracker
	for range b.N {
		tr.Err()
	}
}
