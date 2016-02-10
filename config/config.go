// Package config has some shared configuration helpers
package config

import "encoding/json"

// Args can be used to configure things that don't have mappings until
// checking other options
type Args map[string]interface{}

// ApplyTo unmarshals the options into the given interface for simpler
// configuration.
func (a Args) ApplyTo(i interface{}) (err error) {
	if len(a) > 0 {
		var b []byte
		b, err = json.Marshal(a)
		if err == nil {
			err = json.Unmarshal(b, i)
		}
	}

	return
}
