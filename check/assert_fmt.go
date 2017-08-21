package check

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

// From: https://github.com/stretchr/testify/blob/master/assert/assertions.go

func fmtVals(g, e interface{}) (string, string) {
	if reflect.TypeOf(e) != reflect.TypeOf(g) {
		return fmt.Sprintf("%T(%#v)", g, g), fmt.Sprintf("%T(%#v)", e, e)
	}

	return fmt.Sprintf("%#v", g), fmt.Sprintf("%#v", e)
}

func typeAndKind(v interface{}) (reflect.Type, reflect.Kind) {
	t := reflect.TypeOf(v)
	k := t.Kind()
	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}

	return t, k
}

func diff(g, e interface{}) string {
	if e == nil || g == nil {
		return ""
	}

	et, ek := typeAndKind(e)
	at, _ := typeAndKind(g)

	if et != at {
		return ""
	}

	if ek != reflect.Struct && ek != reflect.Map && ek != reflect.Slice && ek != reflect.Array {
		return ""
	}

	de := spewConfig.Sdump(e)
	dg := spewConfig.Sdump(g)

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
