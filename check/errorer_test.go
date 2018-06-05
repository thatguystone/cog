package check

import (
	"testing"
)

func TestErrorerBasic(t *testing.T) {
	c := New(t)

	er := Errorer{}
	c.NotNil(er.Err())
	c.True(er.Fail())
}

func TestErrorerIgnoreTests(t *testing.T) {
	c := New(t)

	er := Errorer{
		IgnoreTests: true,
	}

	c.Nil(er.Err())
	c.False(er.Fail())
}

func testErrorerOnlyInHere(c *C, er *Errorer) {
	c.NotNil(er.Err())
}

func testErrorerOnlyInNotHere(c *C, er *Errorer) {
	c.Nil(er.Err())
}

func TestErrorerOnlyIn(t *testing.T) {
	c := New(t)

	er := Errorer{
		OnlyIn: []string{"testErrorerOnlyInHere"},
	}

	testErrorerOnlyInHere(c, &er)
	testErrorerOnlyInNotHere(c, &er)
}

func testErrorerSameCodePath(c *C, er *Errorer, fail bool) {
	c.Equal(fail, er.Fail())
}

func TestErrorerSameCodePath(t *testing.T) {
	c := New(t)

	er := Errorer{}
	for i := 0; i < 5; i++ {
		testErrorerSameCodePath(c, &er, i == 0)
	}
}
