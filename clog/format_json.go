package clog

import (
	"encoding/json"

	"github.com/iheartradio/cog/cio/eio"
)

// JSONFormat formats messages as JSON
type JSONFormat struct{}

func init() {
	RegisterFormatter("JSON",
		func(args eio.Args) (Formatter, error) {
			return JSONFormat{}, nil
		})
}

// FormatEntry implements Formatter.FormatEntry
func (JSONFormat) FormatEntry(e Entry) ([]byte, error) {
	return json.Marshal(e)
}

// MimeType implements Formatter.MimeType
func (JSONFormat) MimeType() string {
	return "application/json"
}
