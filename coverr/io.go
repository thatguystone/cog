package coverr

import "io"

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

func (w *writer) Write(b []byte) (int, error) {
	err := w.tr.Err()
	if err != nil {
		return 0, err
	}

	return w.wr.Write(b)
}
