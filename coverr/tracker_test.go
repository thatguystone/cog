package coverr

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestTrackerBasic(t *testing.T) {
	c := check.NewT(t)

	trk := new(Tracker)

	UntilNil(c, 5, func(i int) error {
		return trk.Err()
	})
}

func TestErrCoverage(t *testing.T) {
	_ = Err.Error()
}

func BenchmarkTracker(b *testing.B) {
	var tr Tracker
	for i := 0; i < b.N; i++ {
		tr.Err()
	}
}
