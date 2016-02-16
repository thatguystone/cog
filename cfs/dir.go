package cfs

import (
	"os"
	"path/filepath"
)

// CreateParents creates all parents of the given file
func CreateParents(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0750)
}
