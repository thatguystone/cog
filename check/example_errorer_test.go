package check_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/iheartradio/cog/check"
)

type MockWriter struct {
	check.Errorer
	io.Writer
}

func (mw *MockWriter) Write(b []byte) (int, error) {
	if mw.Fail() {
		return 0, fmt.Errorf("errorer forced to fail")
	}

	return mw.Writer.Write(b)
}

func Example_errorer() {
	buff := &bytes.Buffer{}
	w := MockWriter{Writer: buff}

	// Errors are returned based on the call stack: on the first call from a
	// given stack, Fail() returns true. On any successive calls, Fail() returns
	// false.
	//
	// By doing this, you can test all errors by repeatedly calling the same
	// function until it succeeds. This allows you to test all error pathways
	// and ensure that everything works right. The following code demonstrates
	// this idea.
	err := errors.New("")
	for err != nil {
		_, err = w.Write([]byte("important data"))
		fmt.Println("Error from write:", err)
	}

	// Output:
	// Error from write: errorer forced to fail
	// Error from write: <nil>
}
