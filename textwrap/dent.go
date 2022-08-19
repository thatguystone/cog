// Package textwrap implements python-style text wrapping
package textwrap

import (
	"strings"
	"unicode"
)

func nextLine(s string) (line, rest string) {
	i := strings.IndexByte(s, '\n')
	if i == -1 {
		return s, ""
	}

	return s[:i+1], s[i+1:]
}

// Indent adds prefix to all non-empty lines that do not consist solely of
// whitespace characters.
func Indent(s, prefix string) string {
	return IndentFunc(s, prefix, func(s string) bool {
		return len(strings.TrimSpace(s)) > 0
	})
}

// IndentFunc adds prefix too all lines that where pred(line) returns true.
func IndentFunc(s, prefix string, pred func(line string) bool) string {
	n := strings.Count(s, "\n")

	var b strings.Builder
	b.Grow(len(s) + ((n + 1) * len(prefix)))

	for s != "" {
		var l string
		l, s = nextLine(s)

		if pred(l) {
			b.WriteString(prefix)
		}

		b.WriteString(l)
	}

	return b.String()
}

// Dedent removes any common leading whitespace from every line in s. Entirely
// blank lines are normalized to a single newline character.
func Dedent(s string) string {
	var (
		margin        string
		matches       = 0
		numBlankLines = 0
		numBlankBytes = 0
		orig          = s
	)

	for s != "" {
		var l string
		l, s = nextLine(s)

		trim := strings.TrimLeftFunc(l, unicode.IsSpace)
		if trim == "" {
			numBlankBytes += len(l)
			numBlankLines++
			continue
		}

		indent := l[:len(l)-len(trim)]

		if matches == 0 {
			margin = indent
		}

		matches++

		if strings.HasPrefix(indent, margin) {
			// Indent deeper than margin, no change
			continue
		}

		if strings.HasPrefix(margin, indent) {
			margin = indent
			continue
		}

		for i := 0; i < len(margin) && i < len(indent); i++ {
			if margin[i] != indent[i] {
				margin = margin[:i]
				break
			}
		}
	}

	s = orig

	var b strings.Builder
	b.Grow(len(s) - (matches * len(margin)) - numBlankBytes + numBlankLines)

	for s != "" {
		var l string
		l, s = nextLine(s)

		if len(strings.TrimSpace(l)) == 0 {
			if strings.HasSuffix(l, "\n") {
				b.WriteByte('\n')
			}

			continue
		}

		b.WriteString(strings.TrimPrefix(l, margin))
	}

	return b.String()
}
