package statc

import (
	"time"

	"github.com/thatguystone/cog/cio/eio"
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

// Format implements Formatter.Format
func (LogstashFormat) Format(snap Snapshot) ([]byte, error) {
	now, _ := time.Now().MarshalText()

	snap = append(snap, logstashVersion)
	snap = append(snap, Stat{
		Name: "@timestamp",
		Val:  string(now),
	})

	return JSONFormat{}.Format(snap)
}
