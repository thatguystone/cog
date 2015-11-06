package cfs

import "testing"

func TestChangeExt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		to  string
		out string
	}{
		{
			in:  "test.ext",
			to:  "fun",
			out: "test.fun",
		},
		{
			in:  "test.ext",
			to:  ".fun",
			out: "test.fun",
		},
		{
			in:  "test",
			to:  ".fun",
			out: "test.fun",
		},
		{
			in:  "test.ext",
			to:  "",
			out: "test",
		},
		{
			in:  "test",
			to:  "",
			out: "test",
		},
	}

	for i, test := range tests {
		out := ChangeExt(test.in, test.to)
		if test.out != out {
			t.Errorf("%d: %s != %s", i, test.out, out)
		}
	}
}

func TestDropExt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		out string
	}{
		{
			in:  "test",
			out: "test",
		},
		{
			in:  "test.ext",
			out: "test",
		},
		{
			in:  "test.ext.fun",
			out: "test.ext",
		},
	}

	for i, test := range tests {
		out := DropExt(test.in)
		if test.out != out {
			t.Errorf("%d: %s != %s", i, test.out, out)
		}
	}
}
