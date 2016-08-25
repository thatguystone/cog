package ctime

import (
	"time"

	"github.com/iheartradio/cog"
)

// Backoff is used to limit retries
type Backoff struct {
	Start time.Duration // Amount of time to wait on first failure (default = 5ms)
	Mul   int           // What to multiply start by on successive failures (default = 2)
	Max   time.Duration // Never wait longer than this (default = 2s)
	Exit  *cog.GExit    // To stop waiting early
	curr  time.Duration
}

// Wait adds to the current backoff and sleeps.
//
// Returns if the timeout elapsed and you should try again. If `false`, Exit
// was signaled and you should stop immediately.
func (bo *Backoff) Wait() bool {
	if bo.curr == 0 {
		if bo.Start <= 0 {
			bo.Start = time.Millisecond * 5
		}

		bo.curr = bo.Start
	} else {
		if bo.Mul <= 0 {
			bo.Mul = 2
		}

		bo.curr *= time.Duration(bo.Mul)
	}

	if bo.Max <= 0 {
		bo.Max = time.Second * 2
	}

	if bo.curr > bo.Max {
		bo.curr = bo.Max
	}

	if bo.Exit != nil {
		select {
		case <-bo.Exit.C:
			return false

		case <-time.After(bo.curr):
		}
	} else {
		time.Sleep(bo.curr)
	}

	return true
}

// Reset sets the backoff back to 0
func (bo *Backoff) Reset() {
	bo.curr = 0
}
