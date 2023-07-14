package coverr

import (
	"io"
	"testing"

	"github.com/thatguystone/cog/check"
)

type untilTester struct {
	fatal bool
}

func (untilTester) Helper() {}

func (a *untilTester) Fatal(args ...any) {
	a.fatal = true
}

func TestUntilNil(t *testing.T) {
	c := check.NewT(t)

	c.Run("NoError", func(c *check.T) {
		a := new(untilTester)
		UntilNil(a, 5, func(i int) error {
			return nil
		})

		c.False(a.fatal)
	})

	c.Run("AlwaysErr", func(c *check.T) {
		a := new(untilTester)
		UntilNil(a, 5, func(i int) error { return Err })
		c.True(a.fatal)
	})

	c.Run("NonCoverrErr", func(c *check.T) {
		a := new(untilTester)
		UntilNil(a, 5, func(i int) error {
			if i == 0 {
				return io.EOF
			}

			return nil
		})
		c.True(a.fatal)
	})
}
