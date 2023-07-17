package osx

import (
	"os"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestIsTerminal(t *testing.T) {
	c := check.NewT(t)

	c.Run("Basic", func(c *check.T) {
		f, err := os.Open(os.DevNull)
		c.Must.Nil(err)
		defer f.Close()

		is, err := IsTerminal(f)
		c.Must.Nil(err)
		c.False(is)
	})

	c.Run("InvalidFile", func(c *check.T) {
		var f *os.File

		_, err := IsTerminal(f)
		c.NotNil(err)
	})
}

func TestIsDevNull(t *testing.T) {
	c := check.NewT(t)

	c.Run("Basic", func(c *check.T) {
		f, err := os.Open(os.DevNull)
		c.Must.Nil(err)
		defer f.Close()

		is, err := IsDevNull(f)
		c.Must.Nil(err)
		c.True(is)
	})

	c.Run("InvalidFile", func(c *check.T) {
		var f *os.File

		_, err := IsDevNull(f)
		c.NotNil(err)
	})
}
