package statc

import (
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
	"github.com/iheartradio/cog/cio/eio"
)

type formatErrors struct{}

func init() {
	RegisterFormatter("errors",
		func(eio.Args) (Formatter, error) {
			return formatErrors{}, nil
		})
}

func (formatErrors) FormatSnap(Snapshot) ([]byte, error) {
	return nil, fmt.Errorf("i have issues with that snapshot")
}

func (formatErrors) MimeType() string {
	return "application/errors"
}

func TestFormatErrors(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		RegisterFormatter("json", nil)
	})

	_, err := newFormatter("iDontExist", nil)
	c.Error(err)
}
