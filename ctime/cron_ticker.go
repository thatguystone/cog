package ctime

import (
	"runtime"
	"time"
)

// CronTicker is like a ticker, but it only ticks at specified times. See the
// documentation on UntilPeriod() for when this will fire.
type CronTicker struct {
	C      <-chan time.Time
	exitCh chan struct{}
}

// NewCronTicker creates a new ticker that fires on the given interval
func NewCronTicker(d time.Duration) *CronTicker {
	sendCh := make(chan time.Time, 1)
	pt := &CronTicker{
		C:      sendCh,
		exitCh: make(chan struct{}),
	}

	runtime.SetFinalizer(pt, finalizeCronTicker)

	go runCronTicker(sendCh, pt.exitCh, d)

	return pt
}

func finalizeCronTicker(pt *CronTicker) {
	pt.Stop()
}

// Stop immediately stops this ticker
func (pt *CronTicker) Stop() {
	if pt.exitCh != nil {
		close(pt.exitCh)
		pt.exitCh = nil
		runtime.SetFinalizer(pt, nil)
	}
}

func runCronTicker(
	sendCh chan<- time.Time,
	exitCh <-chan struct{},
	d time.Duration) {

	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	for {
		timer.Reset(UntilPeriod(time.Now(), d))

		select {
		case t := <-timer.C:
			select {
			case sendCh <- t:
			default:
			}

		case <-exitCh:
			return
		}
	}
}
