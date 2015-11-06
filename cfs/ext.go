package cfs

import (
	"path/filepath"
	"strings"
)

// ChangeExt changes the extension on the given file to the given extension. If
// there is no previous extension, the new extension is appended to the path.
func ChangeExt(path, ext string) string {
	rext := filepath.Ext(path)

	if len(ext) > 0 && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return path[0:len(path)-len(rext)] + ext
}

// DropExt is a shortcut for ChangeExt(path, "")
func DropExt(path string) string {
	return ChangeExt(path, "")
}
