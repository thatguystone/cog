package cog

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestMust(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		Must(fmt.Errorf("error"), "nope")
	})

	c.NotPanics(func() {
		Must(nil, "nope")
	})
}

func TestAssert(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		Assert(false, "nope")
	})

	c.NotPanics(func() {
		Assert(true, "nope")
	})
}

func TestBytesMust(t *testing.T) {
	c := check.New(t)

	c.Panics(func() {
		BytesMust(json.Marshal(struct{ Ch chan struct{} }{}))
	})

	c.NotPanics(func() {
		BytesMust(json.Marshal(struct{ S string }{}))
	})
}
