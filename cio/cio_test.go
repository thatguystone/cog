package cio

import (
	"bytes"
	"fmt"
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

	_, err = l.Write([]byte("testing"))
	c.NotNil(err)

	c.Equal(int64(0), l.N)
	c.Equal(128, l.W.(*bytes.Buffer).Len())
}

func TestLimitedWriterWriteError(t *testing.T) {
	c := check.New(t)

	outer := LimitedWriter{
		W: &bytes.Buffer{},
		N: 4,
	}

	inner := LimitedWriter{
		W: &outer,
		N: 128,
	}

	_, err := fmt.Fprintf(&inner, "some long string")
	c.NotNil(err)
}

func TestNopWriteCloser(t *testing.T) {
	check.New(t)

	b := &bytes.Buffer{}

	wc := NopWriteCloser(b)
	wc.Close()
}
