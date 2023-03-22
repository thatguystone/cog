package coverr

import "io"

type reader struct {
	trk *Tracker
	rd  io.Reader
}

// NewReader creates an [io.Reader] that wraps a [Tracker]
func NewReader(trk *Tracker, rd io.Reader) io.Reader {
	return &reader{
		trk: trk,
		rd:  rd,
	}
}

func (rd *reader) Read(b []byte) (int, error) {
	err := rd.trk.Err()
	if err != nil {
		return 0, err
	}

	return rd.rd.Read(b)
}

type writer struct {
	trk *Tracker
	wr  io.Writer
}

// NewWriter creates an [io.Writer] that wraps a [Tracker]
func NewWriter(trk *Tracker, wr io.Writer) io.Writer {
	return &writer{
		trk: trk,
		wr:  wr,
	}
}

func (wr *writer) Write(b []byte) (int, error) {
	err := wr.trk.Err()
	if err != nil {
		return 0, err
	}

	return wr.wr.Write(b)
}
