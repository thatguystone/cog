package bytec

import (
	"bytes"
	"io"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestMultiReaderBasic(t *testing.T) {
	c := check.New(t)

	r := MultiReader(
		[]byte("one"),
		nil,
		[]byte("two"),
		nil,
		[]byte("three"),
		[]byte("4"))

	var b bytes.Buffer

	var err error
	buff := make([]byte, 1)
	for err == nil {
		var n int
		n, err = r.Read(buff)
		if n > 0 {
			b.Write(buff)
		}
	}

	c.Equal(b.String(), "onetwothree4")
}

func TestMultiReaderFullRead(t *testing.T) {
	c := check.New(t)

	r := MultiReader(
		[]byte("one"),
		[]byte("two"),
		[]byte("three"),
		[]byte("4"))

	buff := make([]byte, 128)
	n, err := r.Read(buff)
	c.Equal(err, io.EOF)
	c.Equal(string(buff[:n]), "onetwothree4")
}
