package statc

import (
	"github.com/thatguystone/cog/config"
	"github.com/thatguystone/cog/ctime"
)

// Config sets up stats. It can be read in from a config file to allow for
// simpler configing.
type Config struct {
	// How often to take snapshots. Defaults to 10s.
	SnapshotInterval ctime.HumanDuration

	// Percent of HTTP requests to sample. Defaults to 10%. Range (0-100) for 0% - 100%.
	HTTPSamplePercent int

	// Where stats should be put
	Outputs []OutputConfig
}

// OutputConfig is used to configure outputs
type OutputConfig struct {
	Prod     string      // Which eio Producer to use
	ProdArgs config.Args // Args to pass to the sink
	Fmt      string      // Name of formatter
	FmtArgs  config.Args // Args to pass to the formatter
}

func (cfg *Config) setDefaults() {
	if cfg.SnapshotInterval <= 0 {
		cfg.SnapshotInterval = ctime.Second * 10
	}

	if cfg.HTTPSamplePercent == 0 {
		cfg.HTTPSamplePercent = 10
	}
}