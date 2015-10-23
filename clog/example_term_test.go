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

func init() {
	clog.RegisterFilter("insult", insultFilter{})
}

func Example_terminal() {
	cfg := clog.Config{
		Outputs: map[string]*clog.ConfigOutput{
			"term": {
				Which: "term",
				Level: clog.Info,
				Args: clog.ConfigOutputArgs{
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
				Outputs:       []string{"term"},
				Filters:       []string{"insult"},
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
	// [000000] example_term_test.go:51 : I-polite.module : You're very pretty
	// [000000] example_term_test.go:52 : I-polite.module : I like you
	// [000000] example_term_test.go:57 : E-rude.module : I'm better than you
}
