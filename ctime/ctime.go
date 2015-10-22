// Package ctime implements extra time utilities.
package ctime

import "time"

// UntilPeriod calculates the amount of time until the given period boundary is
// hit.
//
// This is easiest to explain with examples:
//
//      5s: fires every minute at :00, :05, :10, :15, :20, and etc
//     15m: fires every hour at :00, :15, :30, and :45
//     30m: fires every hour at :00 and :30
//      1h: fires every hour at :00
//      2h: fires every 2 hours, on the hour, at 00, 02, 04, 06, 08, 10, 12, 14, 16, 18, 20, and 22
func UntilPeriod(t time.Time, d time.Duration) time.Duration {
	return d - time.Duration(t.UnixNano()%int64(d))
}
