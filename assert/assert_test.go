package assert

import (
	"errors"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestTrue(t *testing.T) {
	c := check.NewT(t)

	c.NotPanics(func() {
		True(true)
	})

	c.Panics(func() {
		True(false)
	})
}

func TestNil(t *testing.T) {
	c := check.NewT(t)

	c.NotPanics(func() {
		Nil(nil)
	})

	c.Panics(func() {
		Nil(errors.New("test"))
	})
}
