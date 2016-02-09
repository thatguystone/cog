package clog

import "encoding/json"

// JSONFormat formats messages as JSON
type JSONFormat struct{}

func init() {
	RegisterFormatter("JSON",
		func(args ConfigArgs) (Formatter, error) {
			return JSONFormat{}, nil
		})
}

// FormatEntry implements Formatter
func (JSONFormat) FormatEntry(e Entry) ([]byte, error) {
	return json.Marshal(e)
}
