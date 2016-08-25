package clog

import (
	"encoding/json"
	"time"

	"github.com/iheartradio/cog/cio/eio"
)

// LogstashFormat formats messages as JSON, with special fields for logstash
type LogstashFormat struct{}

type logstashEntry struct {
	Entry
	Time    time.Time `json:"@timestamp"`
	Host    string    `json:"@host"`
	Version int       `json:"@version"`

	// Hide these fields
	OmitTime *struct{} `json:"time,omitempty"`
	OmitHost *struct{} `json:"host,omitempty"`
}

func init() {
	RegisterFormatter("logstash",
		func(args eio.Args) (Formatter, error) {
			return LogstashFormat{}, nil
		})
}

// FormatEntry implements Formatter.FormatEntry
func (LogstashFormat) FormatEntry(e Entry) ([]byte, error) {
	le := logstashEntry{
		Entry:   e,
		Time:    e.Time,
		Host:    Hostname(),
		Version: 1,
	}

	return json.Marshal(le)
}

// MimeType implements Formatter.MimeType
func (LogstashFormat) MimeType() string {
	return "application/json"
}
