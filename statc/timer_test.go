package statc

import (
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func TestTimerValidate(t *testing.T) {
	c := check.New(t)

	tr := NewTimer(Name{s: "timer"}, 100000)
	c.Equal(tr.sampPercent, 100)

	tr = NewTimer(Name{s: "timer"}, -1)
	c.Equal(tr.sampPercent, 0)
}

func TestTimerTimeFunc(t *testing.T) {
	c := check.New(t)

	tr := NewTimer(Name{s: "timer"}, 100)
	tr.TimeFunc(func() {
		time.Sleep(time.Microsecond)
	})

	c.Len(tr.samples, 1)
}
