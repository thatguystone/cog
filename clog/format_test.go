package clog

import (
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestRegisterFormatterErrors(t *testing.T) {
	c := check.New(t)

	c.Panic(func() {
		RegisterFormatter("JSON",
			func(args ConfigArgs) (Formatter, error) {
				return nil, nil
			})
	})

	_, err := newFormatter(FormatterConfig{Name: "lulz what"})
	c.MustError(err)
}
