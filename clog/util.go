package clog

import (
	"strings"

	"github.com/tchap/go-patricia/patricia"
)

// Normalizes module names and prefixes with dots
func modulePrefix(name string) (patricia.Prefix, string) {
	var pfx patricia.Prefix

	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")

	if len(name) > 0 {
		// Add a dot after each module to ensure that similar prefixes
		// don't match. For example, this prevents "test.some" from
		// being a parent of "test.something".
		pfx = patricia.Prefix(name + ".")
	} else {
		// Patricia needs an array, not a nil val.
		pfx = patricia.Prefix("")
	}

	return pfx, name
}
