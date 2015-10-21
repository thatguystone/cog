package cog

import (
	"fmt"
	"testing"

	"github.com/thatguystone/cog/check"
)

func TestErrorsBasic(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	c.True(es.Empty())
	c.NotError(es.Error())
}

func TestErrorsAdd(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	es.Add(fmt.Errorf("one"))
	c.False(es.Empty())

	c.Error(es.Error())
}

func TestErrorsAddf(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	es.Addf(fmt.Errorf("one"), "some %s stuff", "cool")

	err := es.Error()
	c.MustError(err)
	c.Contains(err.Error(), "some cool stuff: one")
}

func TestErrorsPrefix(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	pes := es.Prefix("test")
	pes.Add(fmt.Errorf("two"))

	err := es.Error()
	c.MustError(err)
	c.Contains(err.Error(), "test: two")
}
