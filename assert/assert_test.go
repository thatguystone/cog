package assert

import (
	"errors"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestTrue(t *testing.T) {
	check.NotPanics(t, func() {
		True(true)
	})

	check.Panics(t, func() {
		True(false)
	})
}

func TestEqual(t *testing.T) {
	check.NotPanics(t, func() {
		Equal(1, 1)
	})

	check.Panics(t, func() {
		Equal(1, 2)
	})
}

func TestNil(t *testing.T) {
	check.NotPanics(t, func() {
		Nil(nil)
	})

	check.Panics(t, func() {
		Nil(errors.New("test"))
	})
}

func TestMust(t *testing.T) {
	check.NotPanics(t, func() {
		v := Must(func() (int, error) { return 1, nil }())
		check.Equal(t, v, 1)
	})

	check.NotPanics(t, func() {
		v := Must(1, nil)
		check.Equal(t, v, 1)
	})

	check.Panics(t, func() {
		Must(1, errors.New("hello"))
	})
}
