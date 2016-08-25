package statc

import (
	"time"

	"github.com/iheartradio/cog/cio/eio"
	"github.com/iheartradio/cog/clog"
)

// LogstashFormat formats snapshots as JSON, ready for shipping to logstash
type LogstashFormat struct{}

var logstashVersion = Stat{
	Name: "@version",
	Val:  int64(1),
}

func init() {
	RegisterFormatter("logstash",
		func(args eio.Args) (Formatter, error) {
			return LogstashFormat{}, nil
		})
}

// FormatSnap implements Formatter.FormatSnap
func (LogstashFormat) FormatSnap(snap Snapshot) ([]byte, error) {
	now, _ := time.Now().MarshalText()

	// Don't modify the existing snapshot
	c := snap.Dup()

	c.add("@version", int64(1))
	c.add("@timestamp", string(now))
	c.add("@host", clog.Hostname())

	return JSONFormat{}.FormatSnap(c)
}

// MimeType implements Formatter.MimeType
func (LogstashFormat) MimeType() string {
	return "application/json"
}
