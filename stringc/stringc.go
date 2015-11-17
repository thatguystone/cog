// Package stringc implements some strings extras
//
// "stringc" is pronounced "strings".
package stringc

import "strings"

// Indent prefixes each line of the given string with the given prefix.
func Indent(prefix, s string) string {
	return prefix + strings.Replace(s, "\n", "\n"+prefix, -1)
}
