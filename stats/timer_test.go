package stats

import (
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func TestTimerValidate(t *testing.T) {
	c := check.New(t)

	tr := NewTimer("timer", 100000)
	c.Equal(tr.sampPercent, 100)

	tr = NewTimer("timer", -1)
	c.Equal(tr.sampPercent, 0)
}

func TestTimerTimeFunc(t *testing.T) {
	c := check.New(t)

	tr := NewTimer("timer", 100)
	tr.TimeFunc(func() {
		time.Sleep(time.Microsecond)
	})

	c.Len(tr.samples, 1)
}
