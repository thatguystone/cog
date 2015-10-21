package cog

import (
	"runtime"
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func TestUntilPeriod(t *testing.T) {
	c := check.New(t)

	tests := []string{
		"5s",
		"10s",
		"15s",
		"15m",
		"1h",
		"32h",
	}

	when := time.Unix(1444852733, 827348)

	for i, test := range tests {
		in, err := time.ParseDuration(test)
		c.MustNotError(err, "error at %d", i)

		res := untilPeriod(when, in)
		at := when.Add(res)
		remain := at.UnixNano() % int64(in)

		c.Equal(0, remain,
			"mismatch at %d, time would be %v (%v)",
			i,
			at,
			at.UnixNano())
	}
}

func TestCronTicker(t *testing.T) {
	check.New(t)

	ticker := NewCronTicker(5 * time.Microsecond)
	defer ticker.Stop()

	for i := 0; i < 3; i++ {
		<-ticker.C
	}
}

func TestCronTickerGC(t *testing.T) {
	check.New(t)

	gcCh := NewCronTicker(5 * time.Microsecond).exitCh

	runtime.GC()

	select {
	case <-gcCh:
	case <-time.After(time.Second):
		t.Fatal("unused ticker not GC'd")
	}
}
