package statc

import (
	"time"

	"github.com/thatguystone/cog/cio/eio"
	"github.com/thatguystone/cog/clog"
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

	snap.add("@version", int64(1))
	snap.add("@timestamp", string(now))
	snap.add("@host", clog.Hostname())

	return JSONFormat{}.FormatSnap(snap)
}

// MimeType implements Formatter.MimeType
func (LogstashFormat) MimeType() string {
	return "application/json"
}
