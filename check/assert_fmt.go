package check

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

// From: https://github.com/stretchr/testify/blob/master/assert/assertions.go

func fmtVals(g, e any) (string, string) {
	if reflect.TypeOf(g) != reflect.TypeOf(e) {
		return fmt.Sprintf("%T(%#v)", g, g), fmt.Sprintf("%T(%#v)", e, e)
	}

	return fmt.Sprintf("%#v", g), fmt.Sprintf("%#v", e)
}

func typeAndKind(v any) (reflect.Type, reflect.Kind) {
	t := reflect.TypeOf(v)
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	return t, k
}

func diff(g, e any) string {
	if g == nil || e == nil {
		return ""
	}

	gt, _ := typeAndKind(g)
	et, ek := typeAndKind(e)

	if gt != et {
		return ""
	}

	if ek != reflect.Struct && ek != reflect.Map && ek != reflect.Slice && ek != reflect.Array {
		return ""
	}

	dg := spewConfig.Sdump(g)
	de := spewConfig.Sdump(e)

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(de),
		FromFile: "Expected",
		B:        difflib.SplitLines(dg),
		ToFile:   "Got",
		Context:  1,
	})

	return diff
}

var spewConfig = spew.ConfigState{
	Indent:                  "    ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
}
