package coverr

import (
	"bytes"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestWriterBasic(t *testing.T) {
	c := check.NewT(t)
	tr := new(Tracker)
	buf := new(bytes.Buffer)

	w := NewWriter(tr, buf)

	UntilNil(c, 10, func(i int) error {
		_, err := w.Write([]byte{1})
		return err
	})

	c.Equal(buf.Bytes(), []byte{1})
}
