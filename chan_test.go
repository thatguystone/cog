package cog

import (
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestNotify(t *testing.T) {
	check.New(t)

	ch := make(chan struct{}, 1)
	Notify(ch)
	<-ch
}
