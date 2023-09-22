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

func TestEqual(t *testing.T) {
	c := check.NewT(t)

	c.NotPanics(func() {
		Equal(1, 1)
	})

	c.Panics(func() {
		Equal(1, 2)
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

func TestMust(t *testing.T) {
	c := check.NewT(t)

	c.NotPanics(func() {
		v := Must(func() (int, error) { return 1, nil }())
		c.Equal(v, 1)
	})

	c.NotPanics(func() {
		v := Must(1, nil)
		c.Equal(v, 1)
	})

	c.Panics(func() {
		Must(1, errors.New("hello"))
	})
}
