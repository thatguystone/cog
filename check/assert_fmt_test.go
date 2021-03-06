package check

import "testing"

func TestFmtVals(t *testing.T) {
	c := New(t)

	a, b := fmtVals(1, 1.0)
	c.Equal(a, "int(1)")
	c.Equal(b, "float64(1)")

	a, b = fmtVals(1, 1)
	c.Equal(a, "1")
	c.Equal(b, "1")
}

func TestDiffCoverage(t *testing.T) {
	c := New(t)

	c.Equal(diff(nil, nil), "")
	c.Equal(diff(1, 1.0), "")
	c.Equal(diff(1, 1), "")

	v := new(int)
	c.Equal(diff(1, v), "")

	c.NotEqual(diff([]string{"test"}, []string{"nope"}), "")
}
