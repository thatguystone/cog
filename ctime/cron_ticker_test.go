package ctime

import (
	"runtime"
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

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
