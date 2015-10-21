package clog

import (
	"fmt"
	"io"
	"strings"
)

// Outputter provides a standard interface for writing log messages
type Outputter interface {
	// An output should typically provide its own message formatting, since many
	// formats are output-specific.
	Formatter

	// Write the formatted message to the backend.
	//
	// This does not implement io.Writer: it's not really a writer, it's an
	// Outputter.
	//
	// If this write fails, a new log entry is be generated that it sent
	// straight to the root logger with level=Error, describing the problem.
	Write([]byte) error

	// When using an external log rotator, this is called to reopen all file
	// handles.
	Reopen() error

	// Get a human-readable representation of this Outputter, for better error
	// reporting. For example, FileOutput would return something like
	// "FileOutput{file:/path/to/file.log}".
	String() string
}

// NewOutputter creates new, configured instances of Outputters.
type NewOutputter func(ConfigOutputArgs) (Outputter, error)

var regdOutputs = map[string]NewOutputter{}

// RegisterOutputter adds a NewOutputter to the list of Outputters
func RegisterOutputter(name string, o NewOutputter) {
	lname := strings.ToLower(name)

	if _, ok := regdOutputs[lname]; ok {
		panic(fmt.Errorf("outputter `%s` already registered", name))
	}

	regdOutputs[lname] = o
}

// DumpKnownOutputs writes all known outputs and their names to the given
// Writer.
func DumpKnownOutputs(w io.Writer) {
	for name, out := range regdOutputs {
		fmt.Fprintf(w, "%s: %v\n", name, out)
	}
}
