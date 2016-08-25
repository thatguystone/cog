package statc

import (
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/ctime"
)

func TestConfigSetDefaults(t *testing.T) {
	c := check.New(t)

	cfg := Config{}
	cfg.setDefaults()

	c.True(cfg.SnapshotInterval > 0)
	c.True(cfg.HTTPSamplePercent > 0)
	c.True(cfg.MemStatsInterval > 0)
	c.True(cfg.MemStatsInterval > cfg.SnapshotInterval)
}

func TestConfigMemStatsIntervalClamp(t *testing.T) {
	c := check.New(t)

	cfg := Config{
		SnapshotInterval: ctime.Second,
		MemStatsInterval: ctime.Millisecond,
	}
	cfg.setDefaults()

	c.True(cfg.HTTPSamplePercent > 0)
	c.Equal(cfg.SnapshotInterval, cfg.MemStatsInterval)
}
