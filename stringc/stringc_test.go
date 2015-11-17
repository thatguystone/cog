package stringc

import (
	"testing"

	"github.com/thatguystone/cog/check"
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
		c.Equal(t.out, Indent(t.p, t.in), "failed at %d", i)
	}
}

func BenchmarkIndent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Indent("\t", "test\nsomething\nlines")
	}
}
