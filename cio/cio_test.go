package cio

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
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
	c.Error(err)

	c.Equal(0, l.N)
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
	c.Error(err)
}
