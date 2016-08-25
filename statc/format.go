package statc

import (
	"fmt"
	"strings"

	"github.com/iheartradio/cog/cio/eio"
)

// A Formatter formats a snapshot
type Formatter interface {
	// Format a snapshot
	FormatSnap(snap Snapshot) ([]byte, error)

	// Get the MimeType of data produced by this Formatter
	MimeType() string
}

// NewFormatter creates a new, configured Formatter.
type NewFormatter func(args eio.Args) (Formatter, error)

var regdFormatters = map[string]NewFormatter{}

// RegisterFormatter adds a Filter to the list of filters
func RegisterFormatter(name string, nf NewFormatter) {
	lname := strings.ToLower(name)

	if _, ok := regdFormatters[lname]; ok {
		panic(fmt.Errorf("formatter `%s` already registered", name))
	}

	regdFormatters[lname] = nf
}

func newFormatter(name string, args eio.Args) (Formatter, error) {
	lname := strings.ToLower(name)
	nf, ok := regdFormatters[lname]
	if !ok {
		return nil, fmt.Errorf("formatter `%s` does not exist", name)
	}

	return nf(args)
}
