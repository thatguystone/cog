package clog

import "encoding/json"

// JSONFormat formats messages as JSON
type JSONFormat struct{}

// FormatEntry implements Formatter
func (JSONFormat) FormatEntry(e Entry) ([]byte, error) {
	return json.Marshal(e)
}
