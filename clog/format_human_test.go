package clog

import (
	"testing"
	"time"

	"github.com/thatguystone/cog/check"
)

func TestHumanFormatCoverage(t *testing.T) {
	check.New(t)

	f := HumanFormat{}
	f.FormatEntry(Entry{
		Time: time.Now(),
		Data: Data{
			"test": 1,
		},
	})

	f.Args.ShortTime = true
	f.FormatEntry(Entry{
		Time: time.Now(),
	})
}
