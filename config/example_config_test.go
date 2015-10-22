package config_test

import (
	"fmt"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/config"
)

type HTTPConfig struct {
	ListenAddress string
	PartyTime     bool
}

func (h *HTTPConfig) Validate(cfg *config.Cfg, es *cog.Errors) {
	if !h.PartyTime {
		es.Add(fmt.Errorf("uhh, it's always party time"))
	}

	cfg.ResolveListen(&h.ListenAddress, es)
}

type MailConfig struct {
	Email   string
	Retries int
}

func (l *MailConfig) Validate(cfg *config.Cfg, es *cog.Errors) {
	if l.Email == "" {
		es.Add(fmt.Errorf("now, I'm not expert, but I think email " +
			"addresses can't be blank"))
	}

	if l.Retries > 5 {
		es.Add(fmt.Errorf("why do you want to retry sending an email %d "+
			"times when the server clearly doesn't want to talk to you?",
			l.Retries))
	}
}

func Example_config() {
	config.Register("ExampleHTTP", func() config.Configer {
		return &HTTPConfig{}
	})

	config.Register("ExampleEmail", func() config.Configer {
		// Set any default values here
		return &MailConfig{
			Retries: 3,
		}
	})

	cfg := config.New()

	c := check.New(nil)
	c.FS.SWriteFile("config.json",
		`{
			"ExampleHTTP": {
				"ListenAddress": ":80",
				"PartyTime": true
			},
			"ExampleEmail": {
				"Email": "email@example.com"
			}
		}`)

	err := cfg.LoadAndValidate(c.FS.Path("config.json"))
	if err != nil {
		panic(fmt.Errorf("failed to config: %v", err))
	}

	fmt.Println(cfg.Modules["ExampleHTTP"].(*HTTPConfig).ListenAddress)
	fmt.Println(cfg.Modules["ExampleEmail"].(*MailConfig).Email)

	// Output:
	// :80
	// email@example.com
}
