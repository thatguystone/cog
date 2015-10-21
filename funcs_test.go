package cog

import (
	"fmt"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestMust(t *testing.T) {
	c := check.New(t)

	c.Panic(func() {
		Must(fmt.Errorf("error"), "nope")
	})

	c.NotPanic(func() {
		Must(nil, "nope")
	})
}
