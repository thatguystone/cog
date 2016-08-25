package clog_test

import (
	"strings"

	"github.com/iheartradio/cog/cio/eio"
	"github.com/iheartradio/cog/clog"
)

// Rejects any messages that might be insulting
type insultFilter struct{}

func (insultFilter) Accept(e clog.Entry) bool {
	return !strings.Contains(strings.ToLower(e.Msg), "i hate you")
}

func (insultFilter) Exit() {
	// This filter has nothing to cleanup, so nothing to do here
}

func init() {
	clog.RegisterFilter("insult",
		func(args eio.Args) (clog.Filter, error) {
			// If args were used here, args.ApplyTo might come in handy
			return insultFilter{}, nil
		})
}

func Example_stdout() {
	cfg := clog.Config{
		Outputs: map[string]*clog.OutputConfig{
			"stdout": {
				Prod: "stdout",
				Fmt:  "human",
				FmtArgs: eio.Args{
					"ShortTime": true,
				},
				Level: clog.Info,
			},
		},

		Modules: map[string]*clog.ModuleConfig{
			"": {
				Outputs: []string{"stdout"},
			},
			"rude.module": {
				Outputs: []string{"stdout"},
				Filters: []clog.FilterConfig{
					clog.FilterConfig{
						Which: "insult",
					},
				},
				DontPropagate: true,
			},
		},
	}

	l, err := clog.New(cfg)
	if err != nil {
		panic(err)
	}

	polite := l.Get("polite.module")
	polite.Info("You're very pretty")
	polite.Info("I like you")

	rude := l.Get("rude.module")
	rude.Info("I hate you")
	rude.Info("You're ugly and I hate you")
	rude.Error("I'm better than you")

	// Output:
	// [000000] I-polite.module : example_stdout_test.go:64 : You're very pretty
	// [000000] I-polite.module : example_stdout_test.go:65 : I like you
	// [000000] E-rude.module : example_stdout_test.go:70 : I'm better than you
}
