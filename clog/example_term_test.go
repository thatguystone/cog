package clog_test

import (
	"strings"

	"github.com/thatguystone/cog/clog"
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
		func(args clog.ConfigArgs) (clog.Filter, error) {
			// If args were used here, args.ApplyTo might come in handy
			return insultFilter{}, nil
		})
}

func Example_terminal() {
	cfg := clog.Config{
		Outputs: map[string]*clog.ConfigOutput{
			"term": {
				Which: "term",
				Level: clog.Info,
				Args: clog.ConfigArgs{
					"ShortTime": true,
					"Stdout":    true,
				},
			},
		},

		Modules: map[string]*clog.ConfigModule{
			"": {
				Outputs: []string{"term"},
			},
			"rude.module": {
				Outputs: []string{"term"},
				Filters: []clog.ConfigFilter{
					clog.ConfigFilter{
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
	// [000000] I-polite.module : example_term_test.go:63 : You're very pretty
	// [000000] I-polite.module : example_term_test.go:64 : I like you
	// [000000] E-rude.module : example_term_test.go:69 : I'm better than you
}
