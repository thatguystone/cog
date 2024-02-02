package coverr

import (
	"bytes"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestReaderBasic(t *testing.T) {
	var (
		tr   = new(Tracker)
		vals = []byte{1, 2, 3, 4}
		buf  = make([]byte, len(vals))
		rd   = NewReader(tr, bytes.NewReader(vals))
	)

	EventuallyNil(t, 10, func(i int) error {
		_, err := rd.Read(buf)
		return err
	})

	check.Equal(t, buf, vals)
}

func TestWriterBasic(t *testing.T) {
	var (
		tr  = new(Tracker)
		buf = new(bytes.Buffer)
		wr  = NewWriter(tr, buf)
	)

	EventuallyNil(t, 10, func(i int) error {
		_, err := wr.Write([]byte{1})
		return err
	})

	check.Equal(t, buf.Bytes(), []byte{1})
}
