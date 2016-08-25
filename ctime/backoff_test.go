package ctime

import (
	"testing"
	"time"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/check"
)

func TestBackoffBasic(t *testing.T) {
	c := check.New(t)

	bo := Backoff{}

	start := time.Now()
	for i := 0; i < 2; i++ {
		bo.Wait()
	}
	end := time.Now()

	c.True(end.Sub(start) > (time.Millisecond * 5))
}

func TestBackoffReset(t *testing.T) {
	c := check.New(t)
	bo := Backoff{
		Start: time.Nanosecond,
		Max:   time.Nanosecond * 5,
	}

	for i := 0; i < 10; i++ {
		bo.Wait()
	}

	c.Equal(bo.curr, bo.Max)
	bo.Reset()
	c.Equal(bo.curr, 0)
}

func TestBackoffCancel(t *testing.T) {
	check.New(t)
	e := cog.NewExit()
	bo := Backoff{
		Start: time.Minute,
		Exit:  e.GExit,
	}

	e.Exit()
	bo.Wait()
}
