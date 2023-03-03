package coverr

import "io"

type reader struct {
	tr *Tracker
	rd io.Reader
}

// NewReader creates an [io.Reader] that wraps a [Tracker]
func NewReader(tr *Tracker, rd io.Reader) io.Reader {
	return &reader{
		tr: tr,
		rd: rd,
	}
}

func (rd *reader) Read(b []byte) (int, error) {
	err := rd.tr.Err()
	if err != nil {
		return 0, err
	}

	return rd.rd.Read(b)
}

type writer struct {
	tr *Tracker
	wr io.Writer
}

// NewWriter creates an [io.Writer] that wraps a [Tracker]
func NewWriter(tr *Tracker, wr io.Writer) io.Writer {
	return &writer{
		tr: tr,
		wr: wr,
	}
}

func (wr *writer) Write(b []byte) (int, error) {
	err := wr.tr.Err()
	if err != nil {
		return 0, err
	}

	return wr.wr.Write(b)
}
