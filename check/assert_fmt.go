package check

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

// From: https://github.com/stretchr/testify/blob/master/assert/assertions.go

func fmtVals(e, g interface{}) (string, string) {
	if reflect.TypeOf(e) != reflect.TypeOf(g) {
		return fmt.Sprintf("%T(%#v)", e, e), fmt.Sprintf("%T(%#v)", g, g)
	}

	return fmt.Sprintf("%#v", e), fmt.Sprintf("%#v", g)
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

func diff(expected interface{}, actual interface{}) string {
	if expected == nil || actual == nil {
		return ""
	}

	et, ek := typeAndKind(expected)
	at, _ := typeAndKind(actual)

	if et != at {
		return ""
	}

	if ek != reflect.Struct && ek != reflect.Map && ek != reflect.Slice && ek != reflect.Array {
		return ""
	}

	e := spewConfig.Sdump(expected)
	a := spewConfig.Sdump(actual)

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(e),
		B:        difflib.SplitLines(a),
		FromFile: "Expected",
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
