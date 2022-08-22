package check

import "testing"

func TestErrorerBasic(t *testing.T) {
	c := NewT(t)

	er := Errorer{}
	c.NotNil(er.Err())
	c.True(er.Fail())
}

func TestErrorerIgnoreTests(t *testing.T) {
	c := NewT(t)

	er := Errorer{
		IgnoreTests: true,
	}

	c.Nil(er.Err())
	c.False(er.Fail())
}

func testErrorerSameCodePath(c *T, er *Errorer, fail bool) {
	c.Equal(fail, er.Fail())
}

func TestErrorerSameCodePath(t *testing.T) {
	c := NewT(t)

	er := Errorer{}
	for i := 0; i < 5; i++ {
		testErrorerSameCodePath(c, &er, i == 0)
	}
}
