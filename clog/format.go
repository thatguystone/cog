package clog

// Formatter formats messages
type Formatter interface {
	// Format a message.
	FormatEntry(Entry) ([]byte, error)
}
