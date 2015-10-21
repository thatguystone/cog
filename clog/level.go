package clog

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Level is a typical log level
type Level int8

const (
	// Debug is the finest level: it contains things that are only useful for
	// debugging
	Debug Level = iota

	// Info is for general information that might be useful during runtime
	Info

	// Warn tells about something that doesn't seem right
	Warn

	// Error is a recoverable runtime error
	Error

	// Panic causes the current goroutine to panic after logging the message
	Panic

	// Fatal causes the entire program to crash after the message is logged
	Fatal
)

// MarshalJSON implements JSON.Marshaler
func (l Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// UnmarshalJSON implements JSON.Unmarshaler
func (l *Level) UnmarshalJSON(data []byte) error {
	var s string

	err := json.Unmarshal(data, &s)
	if err == nil {
		err = l.Parse(s)
	}

	return err
}

// Parse parses a log level from a string.
func (l *Level) Parse(s string) error {
	s = strings.ToLower(s)

	switch {
	case strings.HasPrefix("debug", s):
		*l = Debug
	case strings.HasPrefix("info", s):
		*l = Info
	case strings.HasPrefix("warn", s):
		*l = Warn
	case strings.HasPrefix("error", s):
		*l = Error
	case strings.HasPrefix("panic", s):
		*l = Panic
	case strings.HasPrefix("fatal", s):
		*l = Fatal
	default:
		return fmt.Errorf("unrecognized level: %q", s)
	}

	return nil
}

// Rune gets a single-letter representation of this level
func (l Level) Rune() rune {
	switch l {
	case Debug:
		return 'D'
	case Info:
		return 'I'
	case Warn:
		return 'W'
	case Error:
		return 'E'
	case Panic:
		return 'P'
	case Fatal:
		return 'F'
	default:
		return 'U'
	}
}

// String gets a human-readble version of this level
func (l Level) String() string {
	switch l {
	case Debug:
		return "debug"
	case Info:
		return "info"
	case Warn:
		return "warn"
	case Error:
		return "error"
	case Panic:
		return "panic"
	case Fatal:
		return "fatal"
	default:
		return "unknown"
	}
}
