package clog

import (
	"fmt"
	"io"
	"strings"

	"github.com/thatguystone/cog"
)

// A Filter determines which log entries are allowed through
type Filter interface {
	// Checks whether or not this Entry should be accepted
	Accept(Entry) bool

	// Just in case you need to cleanup when no longer needed
	Exit()
}

// NewFilter creates new, configured Filters.
type NewFilter func(args ConfigArgs) (Filter, error)

type filterSlice []Filter

var regdFilters = map[string]NewFilter{}

// RegisterFilter adds a Filter to the list of filters
func RegisterFilter(name string, nf NewFilter) {
	lname := strings.ToLower(name)

	if _, ok := regdFilters[lname]; ok {
		panic(fmt.Errorf("filter `%s` already registered", name))
	}

	regdFilters[lname] = nf
}

// DumpKnownFilters writes all known filters and their names to the given
// Writer.
func DumpKnownFilters(w io.Writer) {
	for name, f := range regdFilters {
		fmt.Fprintf(w, "%s: %v\n", name, f)
	}
}

func newFilters(lvl Level, cfgs []ConfigFilter) (filts filterSlice, err error) {
	defer func() {
		if err != nil {
			for _, f := range filts {
				f.Exit()
			}
		}
	}()

	f, err := regdFilters[strings.ToLower(lvlFilterName)](ConfigArgs{
		"level": lvl,
	})
	cog.Must(err, "failed to configure built-in %s filter", lvlFilterName)
	filts = append(filts, f)

	for _, cfg := range cfgs {
		nf, ok := regdFilters[strings.ToLower(cfg.Which)]
		if !ok {
			err = fmt.Errorf(`filter "%s" does not exist`, cfg.Which)
			return
		}

		f, err = nf(cfg.Args)
		if err != nil {
			err = fmt.Errorf(`error creating filter "%s": %v`, cfg.Which, err)
			return
		}

		filts = append(filts, f)
	}

	return
}

func (fs filterSlice) accept(e Entry) bool {
	for _, f := range fs {
		if !f.Accept(e) {
			return false
		}
	}

	return true
}
