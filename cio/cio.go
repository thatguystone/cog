// Package cio implements extra io utils
package cio

import "io"

// A LimitedWriter writes to W but limits the amount of data written to just N
// bytes. Each call to Write updates N to reflect the new amount remaining.
type LimitedWriter struct {
	W io.Writer // Underlying Writer
	N int64     // Max bytes remaining
}

func (l *LimitedWriter) Write(b []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}

	if int64(len(b)) > l.N {
		b = b[0:l.N]
		err = io.ErrShortWrite
	}

	n, errw := l.W.Write(b)
	l.N -= int64(n)

	if errw != nil {
		err = errw
	}

	return
}
