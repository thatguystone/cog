package ctime

import (
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

func TestUntilPeriod(t *testing.T) {
	c := check.New(t)

	tests := []string{
		"5s",
		"10s",
		"15s",
		"15m",
		"1h",
		"2h",
		"32h",
	}

	when := time.Unix(1444852733, 827348)

	for i, test := range tests {
		in, err := time.ParseDuration(test)
		c.MustNotError(err, "error at %d", i)

		res := UntilPeriod(when, in)
		at := when.Add(res)
		remain := at.UnixNano() % int64(in)

		c.Equal(0, remain,
			"mismatch at %d, time would be %v (%v)",
			i,
			at,
			at.UnixNano())
	}
}
