package clog

import (
	"bytes"

	"github.com/go-logfmt/logfmt"
	"github.com/iheartradio/cog/cio/eio"
)

// LogFmtFormat formats messages in heroku's LogFmt
type LogFmtFormat struct{}

func init() {
	RegisterFormatter("logfmt",
		func(args eio.Args) (Formatter, error) {
			return LogFmtFormat{}, nil
		})
}

// FormatEntry implements Formatter.FormatEntry
func (LogFmtFormat) FormatEntry(e Entry) ([]byte, error) {
	b := bytes.Buffer{}
	enc := logfmt.NewEncoder(&b)

	err := enc.EncodeKeyval("time", e.Time)

	if err == nil {
		err = enc.EncodeKeyval("src", e.Src)
	}

	if err == nil {
		err = enc.EncodeKeyval("level", e.Level)
	}

	if err == nil {
		err = enc.EncodeKeyval("module", e.Module)
	}

	if err == nil {
		err = enc.EncodeKeyval("msg", e.Msg)
	}

	if err == nil {
		err = enc.EncodeKeyval("host", e.Host)
	}

	for k, v := range e.Data {
		err = enc.EncodeKeyval("data."+k, v)
		if err != nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// MimeType implements Formatter.MimeType
func (LogFmtFormat) MimeType() string {
	return "application/logfmt"
}
