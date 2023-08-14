package check

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
)

// From: https://github.com/stretchr/testify/blob/master/assert/assertions.go

func fmtVals(g, e any) (string, string) {
	gs := fmt.Sprintf("%+v", g)
	es := fmt.Sprintf("%+v", e)

	if reflect.TypeOf(g) != reflect.TypeOf(e) {
		return fmt.Sprintf("%T(%s)", g, gs), fmt.Sprintf("%T(%s)", e, es)
	}

	return gs, es
}

// If fmt.Sprintf("%+v") generally produces a reasonable value
func isFormattable(v any) bool {
	rt := reflect.TypeOf(v)
	if rt == nil {
		return true
	}

	switch rt.Kind() {
	case reflect.Bool:
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Uintptr:
	case reflect.Float32:
	case reflect.Float64:
	case reflect.Complex64:
	case reflect.Complex128:
	case reflect.String:
	default:
		return false
	}

	return true
}

func diff(g, e any) string {
	if isFormattable(g) && isFormattable(e) {
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
