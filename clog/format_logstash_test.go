package clog

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

func TestLogstashFormatOutput(t *testing.T) {
	c := check.New(t)

	f, err := newFormatter("logstash", nil)
	c.MustNotError(err)

	c.Contains(f.MimeType(), "json")

	e := Entry{
		Time: time.Now(),
		Msg:  "stash that log",
		Data: Data{
			"test": 1,
		},
	}

	b, err := f.FormatEntry(e)
	c.MustNotError(err)

	m := map[string]interface{}{}
	err = json.Unmarshal(b, &m)
	c.MustNotError(err)

	tt, err := e.Time.MarshalText()
	c.NotError(err)

	c.Equal(m["@timestamp"], string(tt))
	c.Equal(m["@host"], Hostname())
	c.Equal(m["@version"], 1.0)
	c.Equal(m["time"], nil)
}
