package clog

import "os"

var hostname = ""

// Hostname provides a cached version of os.Hostname() (call that a ton of
// times is probably a bad idea).
func Hostname() string {
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	return hostname
}
