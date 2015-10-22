// Package cfs implements some extra filesystem utilities
package cfs

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// FindDirInParents climbs the directory tree, from the current directory to the
// root, looking for a directory with the given name at each level. If found,
// its absolute path is returned.
func FindDirInParents(dir string) (path string, err error) {
	path, err = os.Getwd()

	if err == nil {
		for path != "/" {
			path = filepath.Join(path, dir)

			exists := false
			exists, err = DirExists(path)

			if err != nil || exists {
				break
			}

			path = filepath.Join(path, "../..")
		}
	}

	if err == nil && path == "/" {
		err = fmt.Errorf("could not find dir %s in parents", dir)
	}

	return
}

func exists(path string) (exists, dir bool, err error) {
	s, err := os.Stat(path)

	if err == nil {
		exists = true
		dir = s.IsDir()
	} else if os.IsNotExist(err) {
		err = nil
	}

	return
}

// FileExists checks if a file exists, taking care of all of the wonkiness
// from os.Stat().
func FileExists(path string) (bool, error) {
	exists, dir, err := exists(path)
	return exists && !dir, err
}

// DirExists checks if a directory exists, taking care of all of the wonkiness
// from os.Stat().
func DirExists(path string) (bool, error) {
	exists, dir, err := exists(path)
	return exists && dir, err
}

// TempFile creates a new temporary file in /tmp/ (or equivalent) with the given
// prefix and, unlike ioutil.TempFile, extension.
func TempFile(prefix, ext string) (f *os.File, err error) {
	dir := os.TempDir()

	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, fmt.Sprintf("%s%d%s",
			prefix,
			rand.Uint32(),
			ext))

		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if !os.IsExist(err) {
			break
		}
	}

	return
}
