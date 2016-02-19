package bytec

import (
	"bytes"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestMultiReader(t *testing.T) {
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
	for err == nil {
		buff := make([]byte, 1)
		_, err = r.Read(buff)
		if err == nil {
			b.Write(buff)
		}
	}

	c.Equal(b.String(), "onetwothree4")
}
