package osx

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal determines if the other end of the given file is a terminal
func IsTerminal(f *os.File) (is bool, err error) {
	c, err := f.SyscallConn()
	if err != nil {
		return
	}

	err = c.Control(func(fd uintptr) {
		is = term.IsTerminal(int(fd))
	})

	return
}

// IsDevNull checks if the given file is connected to [os.DevNull]
func IsDevNull(f *os.File) (is bool, err error) {
	fi, err := f.Stat()
	if err != nil {
		return
	}

	dn, err := os.Open(os.DevNull)
	if err != nil {
		return
	}

	defer dn.Close()

	di, err := dn.Stat()
	if err != nil {
		return
	}

	is = os.SameFile(fi, di)
	return
}
