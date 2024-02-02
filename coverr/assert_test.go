package coverr

import (
	"io"
	"testing"

	"github.com/thatguystone/cog/check"
)

type eventuallyTester struct {
	error bool
	fatal bool
}

func (eventuallyTester) Helper() {}

func (a *eventuallyTester) Error(args ...any) {
	a.error = true
}

func (a *eventuallyTester) Fatal(args ...any) {
	a.fatal = true
}

func TestEventuallyNil(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		tt := new(eventuallyTester)

		EventuallyNil(tt, 5, func(i int) error {
			return nil
		})

		check.False(t, tt.error)
	})

	t.Run("AlwaysErr", func(t *testing.T) {
		tt := new(eventuallyTester)
		EventuallyNil(tt, 5, func(i int) error { return Err })
		check.True(t, tt.error)
	})

	t.Run("NonCoverrErr", func(t *testing.T) {
		tt := new(eventuallyTester)

		EventuallyNil(tt, 5, func(i int) error {
			if i == 0 {
				return io.EOF
			}

			return nil
		})

		check.True(t, tt.error)
	})
}

func TestMustEventuallyNil(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		tt := new(eventuallyTester)

		MustEventuallyNil(tt, 5, func(i int) error {
			return nil
		})

		check.False(t, tt.fatal)
	})

	t.Run("AlwaysErr", func(t *testing.T) {
		tt := new(eventuallyTester)
		MustEventuallyNil(tt, 5, func(i int) error { return Err })
		check.True(t, tt.fatal)
	})

	t.Run("NonCoverrErr", func(t *testing.T) {
		tt := new(eventuallyTester)

		MustEventuallyNil(tt, 5, func(i int) error {
			if i == 0 {
				return io.EOF
			}

			return nil
		})

		check.True(t, tt.fatal)
	})
}
