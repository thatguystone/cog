// Package chlog provides clog-based logging for testing
package chlog

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/clog"
)

// New creates a new TestLog that outputs all log messages to the given TB. This
// logs at level Debug by default.
func New(tb testing.TB) (*check.C, *clog.Log) {
	c := check.New(tb)

	lcfg := clog.Config{
		Outputs: map[string]*clog.ConfigOutput{
			"testlog": {
				Which: "testlog",
				Level: clog.Debug,
				Args: clog.ConfigOutputArgs{
					"log": c,
				},
			},
		},
		Modules: map[string]*clog.ConfigModule{
			"": {
				Outputs: []string{"testlog"},
				Level:   clog.Debug,
			},
		},
	}

	log, err := clog.New(lcfg)
	c.MustNotError(err)

	return c, log
}
