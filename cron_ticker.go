package cog

import (
	"runtime"
	"time"
)

// CronTicker is like a ticker, but it only ticks at specified times.
//
// When created with an interval of 15m, the ticker will only fire on :00, :15,
// :30, and :45. When created with an interval of 5m, it will fire at :00, :05,
// :10, :15, and so on.
type CronTicker struct {
	C      <-chan time.Time
	exitCh chan struct{}
}

func untilPeriod(t time.Time, d time.Duration) time.Duration {
	return d - time.Duration(t.UnixNano()%int64(d))
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
	}
}

func runCronTicker(
	sendCh chan<- time.Time,
	exitCh <-chan struct{},
	d time.Duration) {

	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	for {
		timer.Reset(untilPeriod(time.Now(), d))

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
