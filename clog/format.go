package clog

import (
	"fmt"
	"strings"

	"github.com/thatguystone/cog/config"
)

// Formatter formats messages
type Formatter interface {
	// Format a message.
	FormatEntry(Entry) ([]byte, error)
}

// NewFormatter creates a new, configured Formatter.
type NewFormatter func(args config.Args) (Formatter, error)

var regdFormatters = map[string]NewFormatter{}

// RegisterFormatter adds a Filter to the list of filters
func RegisterFormatter(name string, nf NewFormatter) {
	lname := strings.ToLower(name)

	if _, ok := regdFormatters[lname]; ok {
		panic(fmt.Errorf("formatter `%s` already registered", name))
	}

	regdFormatters[lname] = nf
}

func newFormatter(cfg FormatterConfig) (Formatter, error) {
	lname := strings.ToLower(cfg.Name)
	nf, ok := regdFormatters[lname]
	if !ok {
		return nil, fmt.Errorf("formatter %s does not exist", cfg.Name)
	}

	return nf(cfg.Args)
}
