package clog

import (
	"fmt"
	"io"
	"strings"
)

// A Filter determines which log entries are allowed through
type Filter interface {
	// Checks whether or not this Entry should be accepted
	Accept(Entry) bool
}

type filters struct {
	lvl Level
	s   []Filter
}

var regdFilters = map[string]Filter{}

// RegisterFilter adds a Filter to the list of filters
func RegisterFilter(name string, f Filter) {
	lname := strings.ToLower(name)

	if _, ok := regdFilters[lname]; ok {
		panic(fmt.Errorf("filter `%s` already registered", name))
	}

	regdFilters[lname] = f
}

// DumpKnownFilters writes all known filters and their names to the given
// Writer.
func DumpKnownFilters(w io.Writer) {
	for name, f := range regdFilters {
		fmt.Fprintf(w, "%s: %v\n", name, f)
	}
}

func newFilters(lvl Level, ss []string) (*filters, error) {
	fs := &filters{
		lvl: lvl,
	}

	for _, s := range ss {
		f, ok := regdFilters[strings.ToLower(s)]
		if !ok {
			return nil, fmt.Errorf(`filter "%s" does not exist`, s)
		}

		fs.s = append(fs.s, f)
	}

	return fs, nil
}

func (fs *filters) levelEnabled(lvl Level) bool {
	return lvl >= fs.lvl
}

func (fs *filters) accept(e Entry) bool {
	if !fs.levelEnabled(e.Level) {
		return false
	}

	for _, f := range fs.s {
		if !f.Accept(e) {
			return false
		}
	}

	return true
}
