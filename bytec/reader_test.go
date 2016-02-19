package bytec

import (
	"io/ioutil"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestMultiReader(t *testing.T) {
	c := check.New(t)

	r := MultiReader(
		[]byte("one"),
		nil,
		[]byte("two"),
		nil,
		[]byte("three"),
		[]byte("4"))

	b, err := ioutil.ReadAll(r)
	c.MustNotError(err)

	c.Equal(string(b), "onetwothree4")
}
