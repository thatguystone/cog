package statc

import "strings"

// JoinPath creates a stats path from multiple path elements. The parent element
// is left unescaped to allow for tacking on child paths.
func JoinPath(parent string, parts ...string) string {
	for i, part := range parts {
		parts[i] = EscapePath(part)
	}

	path := parent + "." + strings.Join(parts, ".")
	return CleanPath(path)
}

// JoinNoEscape joins all parts together without escaping each part. This
// means that parts may contain ".".
func JoinNoEscape(parts ...string) string {
	return CleanPath(strings.Join(parts, "."))
}

// CleanPath cleans extra dots and such from the path
func CleanPath(path string) string {
	for strings.Contains(path, "..") {
		path = strings.Replace(path, "..", ".", -1)
	}

	path = strings.TrimPrefix(path, ".")
	path = strings.TrimSuffix(path, ".")

	return path
}

// EscapePath escapes all reserved characters in the given path. For example,
// if you have a path like "url/img.jpg", the "." presents a special problem,
// so the path becomes "url/img_jpg".
//
// This might do more escapes in the future, so use this instead of
// strings.Replace().
func EscapePath(path string) string {
	return strings.Replace(path, ".", "_", -1)
}
