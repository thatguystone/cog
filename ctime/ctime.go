// Package ctime implements extra time utilities.
package ctime

import "time"

// UntilPeriod calculates the amount of time until the given period boundary is
// hit. That is, for example, if the period is "15m", and the given time is
// 13:03, it would return "12m", meaning 12 minutes until it's "13:15". If it's
// "30m", and the time is 13:15, it would return "15m", meaning 15 minutes until
// "13:30".
func UntilPeriod(t time.Time, d time.Duration) time.Duration {
	return d - time.Duration(t.UnixNano()%int64(d))
}
