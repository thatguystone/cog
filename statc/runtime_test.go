package statc

import (
	"testing"
	"time"

	"github.com/iheartradio/cog/ctime"
)

func TestRuntimeBasic(t *testing.T) {
	c, st := newTest(t, &Config{
		SnapshotInterval: ctime.Millisecond,
		MemStatsInterval: ctime.Millisecond,
	})
	defer st.exit.Exit()

	// There aren't really any values that can be tested here since none of the
	// runtime stats are deterministic. So let's just settle for stats being
	// set!
	c.Until(time.Second, func() bool {
		snap := st.snapshot()
		return snap.Get(st.Names("runtime", "mem", "heap", "alloc")).Val.(int64) > 0
	})
}
