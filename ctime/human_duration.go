package ctime

import (
	"encoding/json"
	"time"
)

// HumanDuration wraps time.Duration to provide human-usable duration parsing
// from JSON
type HumanDuration struct {
	time.Duration
}

// MarshalJSON is for JSON
func (d HumanDuration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON is for JSON
func (d *HumanDuration) UnmarshalJSON(b []byte) (err error) {
	var val time.Duration

	if b[0] == '"' {
		sd := string(b[1 : len(b)-1])
		val, err = time.ParseDuration(sd)
	} else {
		var i int64
		i, err = json.Number(string(b)).Int64()
		val = time.Duration(i)
	}

	d.Duration = val

	return
}
