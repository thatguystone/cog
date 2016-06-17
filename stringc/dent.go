// Package stringc implements some strings extras
//
// "stringc" is pronounced "strings".
package stringc

import (
	"bytes"
	"strings"
	"unicode"
)

// Indent prefixes each line of the given string with the given prefix.
func Indent(s, prefix string) string {
	return prefix + strings.Replace(s, "\n", "\n"+prefix, -1)
}

// Dedent is the opposite of indent: it removes the given number of indents
// from each line. The indent style and width is assumed to be the total
// whitespace before the first non-whitespace character on the first line.
//
// This is a shortcut from DedentPrefix(s, "", n)
func Dedent(s string, n int) string {
	return DedentPrefix(s, "", n)
}

// DedentPrefix is the opposite of indent: it removes the given prefix n
// number of times from each line.
func DedentPrefix(s, prefix string, n int) (out string) {
	if n <= 0 {
		return s
	}

	b := bytes.NewBufferString(s)

	for i := 0; b.Len() > 0; i++ {
		l, _ := b.ReadString('\n')

		if i == 0 && prefix == "" {
			tl := strings.TrimLeftFunc(l, unicode.IsSpace)
			prefix = l[:len(l)-len(tl)]
		}

		for j := 0; j < n && strings.HasPrefix(l, prefix); j++ {
			l = l[len(prefix):]
		}

		out += l
	}

	return
}
