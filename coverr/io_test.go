package coverr

import (
	"bytes"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestReaderBasic(t *testing.T) {
	c := check.NewT(t)

	var (
		tr   = new(Tracker)
		vals = []byte{1, 2, 3, 4}
		buf  = make([]byte, len(vals))
		rd   = NewReader(tr, bytes.NewReader(vals))
	)

	UntilNil(c, 10, func(i int) error {
		_, err := rd.Read(buf)
		return err
	})

	c.Equal(buf, vals)
}

func TestWriterBasic(t *testing.T) {
	c := check.NewT(t)

	var (
		tr  = new(Tracker)
		buf = new(bytes.Buffer)
		wr  = NewWriter(tr, buf)
	)

	UntilNil(c, 10, func(i int) error {
		_, err := wr.Write([]byte{1})
		return err
	})

	c.Equal(buf.Bytes(), []byte{1})
}
