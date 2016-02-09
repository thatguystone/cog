package stats

import "strings"

// Join creates a stats path from multiple path elements
func Join(parts ...string) string {
	path := strings.Join(parts, ".")
	for strings.Contains(path, "..") {
		path = strings.Replace(path, "..", ".", -1)
	}

	path = strings.TrimPrefix(path, ".")
	path = strings.TrimSuffix(path, ".")

	return path
}
