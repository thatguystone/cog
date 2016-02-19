package clog

import (
	"encoding/json"
	"time"

	"github.com/thatguystone/cog/cio/eio"
)

// LogstashFormat formats messages as JSON, with special fields for logstash
type LogstashFormat struct{}

type logstashEntry struct {
	Entry
	Time    time.Time `json:"@timestamp"`
	Version int       `json:"@version"`

	// Hide time field in Entry
	OmitTime *struct{} `json:"time,omitempty"`
}

func init() {
	RegisterFormatter("logstash",
		func(args eio.Args) (Formatter, error) {
			return LogstashFormat{}, nil
		})
}

// FormatEntry implements Formatter
func (LogstashFormat) FormatEntry(e Entry) ([]byte, error) {
	le := logstashEntry{
		Entry:   e,
		Time:    e.Time,
		Version: 1,
	}

	return json.Marshal(le)
}
