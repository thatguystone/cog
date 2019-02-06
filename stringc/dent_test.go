package stringc_test

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/stringc"
)

func TestIndent(t *testing.T) {
	c := check.New(t)

	tests := []struct {
		p   string
		in  string
		out string
	}{
		{
			p:   "\t",
			in:  "test\nsomething\nlines",
			out: "\ttest\n\tsomething\n\tlines",
		},
		{
			p:   "   ",
			in:  "test\nsomething\nlines",
			out: "   test\n   something\n   lines",
		},
		{
			p:   " * ",
			in:  "test\nsomething\nlines",
			out: " * test\n * something\n * lines",
		},
	}

	for i, t := range tests {
		c.Equalf(t.out, stringc.Indent(t.in, t.p), "failed at %d", i)
	}
}

func TestDedent(t *testing.T) {
	c := check.New(t)
	c.Equal("test\nlines", stringc.Dedent("  test\n  lines", 1))
}

func TestDedentPrefix(t *testing.T) {
	c := check.New(t)

	tests := []struct {
		p   string
		in  string
		out string
		n   int
	}{
		{
			p:   "",
			in:  "  test\n  something\n  lines",
			out: "test\nsomething\nlines",
			n:   1,
		},
		{
			p:   "   ",
			in:  "    test\n    something\n    lines",
			out: " test\n something\n lines",
			n:   1,
		},
		{
			p:   " * ",
			in:  " * test\n * something\n * lines",
			out: "test\nsomething\nlines",
			n:   1,
		},
		{
			p:   " ",
			in:  "   test\n   something\n   lines",
			out: " test\n something\n lines",
			n:   2,
		},
		{
			p:   "",
			in:  "   test\n   something\n   lines",
			out: "   test\n   something\n   lines",
			n:   0,
		},
		{
			p:   "",
			in:  "\n\n\n\n   test\n   something\n   lines",
			out: "test\nsomething\nlines",
			n:   1,
		},
	}

	for i, t := range tests {
		c.Equalf(t.out, stringc.DedentPrefix(t.in, t.p, t.n), "failed at %d", i)
	}
}

func BenchmarkIndent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stringc.Indent("\t", "test\nsomething\nlines")
	}
}

func BenchmarkDedent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stringc.DedentPrefix("\t\t\ttest\n\t\t\tsomething\n\t\t\tlines", "\t", 3)
	}
}
