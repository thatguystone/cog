// Package chlog provides clog-based logging for testing
//
// Usage is really simple:
//
//    import "github.com/iheartradio/cog/check/chlog"
//    func TestStuff(t *testing.T) {
//        log := chlog.New(t)
//    }
package chlog

import (
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
	"github.com/iheartradio/cog/clog"
)

// New creates a new TestLog that outputs all log messages to the given TB. This
// logs at level Debug by default.
func New(tb testing.TB) (*check.C, *clog.Ctx) {
	c := check.New(tb)

	log, err := clog.New(Config(c))
	c.MustNotError(err)

	return c, log
}

// Config gets logging config for testing
func Config(c *check.C) clog.Config {
	return clog.Config{
		Outputs: map[string]*clog.OutputConfig{
			"testlog": {
				Prod: "testlog",
				ProdArgs: eio.Args{
					"log": c,
				},
				Fmt:   "human",
				Level: clog.Debug,
			},
		},
		Modules: map[string]*clog.ModuleConfig{
			"": {
				Outputs: []string{"testlog"},
				Level:   clog.Debug,
			},
		},
	}
}
