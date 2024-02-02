package osx

import (
	"os"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestIsTerminal(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		f, err := os.Open(os.DevNull)
		check.MustNil(t, err)
		defer f.Close()

		is, err := IsTerminal(f)
		check.MustNil(t, err)
		check.False(t, is)
	})

	t.Run("InvalidFile", func(t *testing.T) {
		var f *os.File

		_, err := IsTerminal(f)
		check.NotNil(t, err)
	})
}

func TestIsDevNull(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		f, err := os.Open(os.DevNull)
		check.MustNil(t, err)
		defer f.Close()

		is, err := IsDevNull(f)
		check.MustNil(t, err)
		check.True(t, is)
	})

	t.Run("InvalidFile", func(t *testing.T) {
		var f *os.File

		_, err := IsDevNull(f)
		check.NotNil(t, err)
	})
}
