// Package cfs implements some extra filesystem utilities
package cfs

import (
	"os"
)

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
