package clog

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

var (
	levels         = []Level{Debug, Info, Warn, Error, Panic}
	badLevel Level = 127
)

func TestLevelParse(t *testing.T) {
	c := check.New(t)

	var ol Level

	for _, l := range levels {
		err := ol.Parse(l.String())
		c.NotError(err, "failed to parse level for %s", l)
		c.Equal(l, ol)
	}

	c.Error(ol.Parse(badLevel.String()))
}

func TestLevelRune(t *testing.T) {
	check.New(t)

	for _, l := range levels {
		l.Rune()
	}

	badLevel.Rune()
}
