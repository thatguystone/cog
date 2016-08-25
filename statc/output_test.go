package statc

import (
	"strings"
	"testing"
	"time"

	"github.com/iheartradio/cog/clog"
	"github.com/iheartradio/cog/ctime"
)

func TestOutputBasic(t *testing.T) {
	c, st := newTest(t, nil)
	defer st.exit.Exit()

	st.NewCounter("cnt", true).Add(100)
	st.NewTimer("timer", 100).Add(time.Millisecond)

	conts := ""
	c.Until(time.Second, func() bool {
		conts = c.FS.SReadFile("stats")
		return strings.Count(conts, "\n") > 1
	})

	c.Contains(conts, `"cnt":100`)
	c.Contains(conts, `"count":1`)
}

func testOutputErrors(t *testing.T, errStr string, ocfg OutputConfig) {
	c, st := newTest(t, &Config{
		SnapshotInterval:  ctime.Millisecond,
		HTTPSamplePercent: 100,
		Outputs: []OutputConfig{
			ocfg,
		},
	})
	defer st.exit.Exit()

	err := st.log.Reconfigure(clog.Config{
		File: c.FS.Path("log"),
	})
	c.MustNotError(err)

	conts := ""
	c.Until(time.Second, func() bool {
		conts = c.FS.SReadFile("log")
		return strings.Count(conts, "\n") > 1
	})

	c.Contains(conts, errStr)
}

func TestOutputFormatErrors(t *testing.T) {
	testOutputErrors(t,
		"format error: i have issues with that snapshot",
		OutputConfig{
			Prod: "blackhole",
			Fmt:  "errors",
		})
}

func TestOutputProducerErrors(t *testing.T) {
	testOutputErrors(t,
		"producer error: i only produce errors",
		OutputConfig{
			Prod: "test_errors",
			Fmt:  "json",
		})
}
