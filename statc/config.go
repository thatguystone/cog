package statc

import (
	"github.com/iheartradio/cog/cio/eio"
	"github.com/iheartradio/cog/ctime"
)

// Config sets up stats. It can be read in from a config file to allow for
// simpler configing.
type Config struct {
	// How often to take snapshots. Defaults to 10s.
	SnapshotInterval ctime.HumanDuration

	// Percent of HTTP requests to sample. Defaults to 10%. Range (0-100) for
	// 0% - 100%.
	HTTPSamplePercent int

	// If memory stats should be included in reported statistics. Defaults to
	// 60s; gathering these stats stops the world, so don't do this too often.
	// Set to <0 to disable.
	MemStatsInterval ctime.HumanDuration

	// StatusKey is the key used to secure the HTTPMuxer's /_status endpoint
	StatusKey string

	// Where stats should be put
	Outputs []OutputConfig

	// For testing
	disableRuntimeStats bool
}

// OutputConfig is used to configure outputs
type OutputConfig struct {
	Prod     string   // Which eio Producer to use
	ProdArgs eio.Args // Args to pass to the sink
	Fmt      string   // Name of formatter
	FmtArgs  eio.Args // Args to pass to the formatter
}

func (cfg *Config) setDefaults() {
	if cfg.SnapshotInterval <= 0 {
		cfg.SnapshotInterval = ctime.Second * 10
	}

	if cfg.HTTPSamplePercent == 0 {
		cfg.HTTPSamplePercent = 10
	}

	if cfg.MemStatsInterval == 0 {
		cfg.MemStatsInterval = ctime.Second * 60
	}

	// That would just be completely pointless
	if cfg.MemStatsInterval > 0 && cfg.MemStatsInterval < cfg.SnapshotInterval {
		cfg.MemStatsInterval = cfg.SnapshotInterval
	}
}
