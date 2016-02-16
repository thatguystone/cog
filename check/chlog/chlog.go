// Package chlog provides clog-based logging for testing
//
// Usage is really simple:
//
//    import "github.com/thatguystone/cog/check/chlog"
//    func TestStuff(t *testing.T) {
//        log := chlog.New(t)
//    }
package chlog

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/cio/eio"
	"github.com/thatguystone/cog/clog"
)

// New creates a new TestLog that outputs all log messages to the given TB. This
// logs at level Debug by default.
func New(tb testing.TB) (*check.C, *clog.Log) {
	c := check.New(tb)

	lcfg := clog.Config{
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

	log, err := clog.New(lcfg)
	c.MustNotError(err)

	return c, log
}
