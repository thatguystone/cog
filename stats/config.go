package stats

import "github.com/thatguystone/cog/ctime"

// Config sets up stats. It can be read in from a config file to allow for
// simpler configing.
type Config struct {
	// How often to flush stats
	FlushInterval ctime.HumanDuration

	// Percent of HTTP requests to sample
	HTTPSamplePercent int
}
