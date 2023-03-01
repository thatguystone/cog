package coverr_test

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/thatguystone/cog/coverr"
)

// Typically, you'd use [coverr.NewWriter] instead
type MockWriter struct {
	buf bytes.Buffer
	tr  *coverr.Tracker
}

func (w *MockWriter) Write(b []byte) (int, error) {
	err := w.tr.Err()
	if err != nil {
		return 0, err
	}

	return w.buf.Write(b)
}

func Example_mock() {
	var tr coverr.Tracker

	for {
		w := MockWriter{tr: &tr}
		err := json.NewEncoder(&w).Encode(map[string]string{
			"test": "test",
		})
		if err == nil {
			break
		}

		if !errors.Is(err, coverr.Err) {
			panic(err)
		}
	}
}
