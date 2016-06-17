package stringc

import "strings"

// TrimLines removes empty lines from both the left and right of s.
func TrimLines(s string) string {
	return TrimRightLines(TrimLeftLines(s))
}

// TrimLeftLines removes empty lines from the left of s.
func TrimLeftLines(s string) string {
	for len(s) > 0 {
		i := strings.Index(s, "\n")
		if i == -1 {
			break
		}

		l := strings.TrimSpace(s[:i+1])
		if len(l) > 0 {
			break
		}

		s = s[i+1:]
	}

	return trimEmpty(s)
}

// TrimRightLines removes empty lines from the right of s.
func TrimRightLines(s string) string {
	for len(s) > 0 {
		i := strings.LastIndex(s, "\n")
		if i == -1 {
			break
		}

		l := strings.TrimSpace(s[i:])
		if len(l) > 0 {
			break
		}

		s = s[:i]
	}

	return trimEmpty(s)
}

func trimEmpty(s string) string {
	ss := strings.TrimSpace(s)
	if len(ss) == 0 {
		s = ss
	}

	return s
}
