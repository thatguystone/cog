package cio

import (
	"bytes"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestLimitedWriter(t *testing.T) {
	c := check.New(t)

	l := LimitedWriter{
		W: &bytes.Buffer{},
		N: 128,
	}

	var err error
	for err == nil {
		_, err = l.Write([]byte("testing"))
	}

	c.Equal(0, l.N)
	c.Equal(128, l.W.(*bytes.Buffer).Len())
}
