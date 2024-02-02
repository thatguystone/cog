package textwrap_test

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/textwrap"
)

func TestIndentBasic(t *testing.T) {
	tests := []struct {
		p   string
		in  string
		out string
	}{
		{
			p:   "",
			in:  "",
			out: "",
		},
		{
			p:   "",
			in:  "test\nsomething\nlines",
			out: "test\nsomething\nlines",
		},
		{
			p:   "\t",
			in:  "",
			out: "",
		},
		{
			p:   "\t",
			in:  "\n",
			out: "\n",
		},
		{
			p:   "\t",
			in:  "\na\n",
			out: "\n\ta\n",
		},
		{
			p:   "\t",
			in:  "test",
			out: "\ttest",
		},
		{
			p:   "\t",
			in:  "a\n\nb",
			out: "\ta\n\n\tb",
		},
		{
			p:   "\t",
			in:  "a\n\nb\n",
			out: "\ta\n\n\tb\n",
		},
		{
			p:   "\t",
			in:  "test\n",
			out: "\ttest\n",
		},
		{
			p:   "\t",
			in:  "test\nsomething\nlines",
			out: "\ttest\n\tsomething\n\tlines",
		},
		{
			p:   "\t",
			in:  "test\nsomething\nlines\n",
			out: "\ttest\n\tsomething\n\tlines\n",
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

	for i, test := range tests {
		check.Equalf(t, test.out, textwrap.Indent(test.in, test.p), "failed at %d", i)

		n := testing.AllocsPerRun(1_000, func() {
			textwrap.Dedent(test.in)
		})
		check.Truef(t, n <= 1.0, "i=%d n=%f", i, n)
	}
}

func TestDedentBasic(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "",
			out: "",
		},
		{
			in:  "  ",
			out: "",
		},
		{
			in:  "  \n",
			out: "\n",
		},
		{
			in:  "  \n  ",
			out: "\n",
		},
		{
			in:  "  \n  \n",
			out: "\n\n",
		},
		{
			in:  "  \n  \n  a\n b",
			out: "\n\n a\nb",
		},
		{
			in:  "  \n    \n  a\n b\n",
			out: "\n\n a\nb\n",
		},
		{
			in:  " a\n   b\n",
			out: "a\n  b\n",
		},
		{
			in:  "   a\n b\n",
			out: "  a\nb\n",
		},
		{
			in:  "\t\ta\n\tb\n",
			out: "\ta\nb\n",
		},
		{
			in:  "  a\n \tb",
			out: " a\n\tb",
		},
		{
			in:  "    a\n\t\t\t\tb",
			out: "    a\n\t\t\t\tb",
		},
		{
			in:  "    a\n  \t b",
			out: "  a\n\t b",
		},
		{
			in:  "a",
			out: "a",
		},
		{
			in:  "test\nsomething\nlines",
			out: "test\nsomething\nlines",
		},
		{
			in:  "test\nsomething\nlines\n",
			out: "test\nsomething\nlines\n",
		},
		{
			in:  "\ttest\n\tsomething\n\tlines",
			out: "test\nsomething\nlines",
		},
		{
			in:  "test\n\tsomething\n\tlines",
			out: "test\n\tsomething\n\tlines",
		},
		{
			in:  "test\n\tsomething\n\tlines\n",
			out: "test\n\tsomething\n\tlines\n",
		},
		{
			in:  "test\n\t  \n\tlines",
			out: "test\n\n\tlines",
		},
	}

	for i, test := range tests {
		check.Equalf(t, test.out, textwrap.Dedent(test.in), "i=%d", i)

		n := testing.AllocsPerRun(1_000, func() {
			textwrap.Dedent(test.in)
		})
		check.Truef(t, n <= 1.0, "i=%d n=%f", i, n)
	}
}

func BenchmarkIndent(b *testing.B) {
	for range b.N {
		textwrap.Indent("test\nsomething\nlines", "\t")
	}
}

func BenchmarkDedent(b *testing.B) {
	for range b.N {
		textwrap.Dedent("\t\t\ttest\n\t\t\tsomething\n\t\t\tlines")
	}
}
