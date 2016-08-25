package cog

import (
	"fmt"
	"testing"

	"github.com/iheartradio/cog/check"
)

func TestErrorsBasic(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	c.True(es.Empty())
	c.NotError(es.Error())
}

func TestErrorsAddAndReset(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	es.Add(fmt.Errorf("one"))
	c.False(es.Empty())
	c.Error(es.Error())

	es.Reset()
	c.True(es.Empty())
	c.NotError(es.Error())
}

func TestErrorsAddf(t *testing.T) {
	c := check.New(t)

	es := Errors{}
	es.Addf(fmt.Errorf("one"), "some %s stuff", "cool")

	err := es.Error()
	c.MustError(err)
	c.Contains(err.Error(), "some cool stuff: one")
}

func TestErrorsDrain(t *testing.T) {
	c := check.New(t)

	errs := make(chan error, 16)
	for i := 0; i < cap(errs); i++ {
		errs <- fmt.Errorf("err %d", i)
	}
	close(errs)

	es := Errors{}
	es.Drain(errs)
	c.Len(*es.errs, cap(errs))
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

func ExampleErrors() {
	es := Errors{}

	// Nothing added
	es.Add(nil)
	es.Addf(nil, "something describing the error: %d", 123)
	fmt.Println("no errors:", es.Error())

	// Errors logged
	es.Add(fmt.Errorf("some error"))
	es.Addf(fmt.Errorf("another error"), "something describing the error: %d", 123)
	fmt.Println("errors:", es.Error())

	// Output:
	// no errors: <nil>
	// errors: some error
	// something describing the error: 123: another error
}
