// Package assert implements assertions for assumed behavior
package assert

import "fmt"

// True ensures that the condition is true
func True(v bool) {
	if !v {
		panic("assert error")
	}
}

// Equal ensures that two values are equal
func Equal[T comparable](a, b T) {
	if a != b {
		panic(fmt.Errorf("%v != %v", a, b))
	}
}

// Nil asserts that v must be nil, or it panics
func Nil(v any) {
	if v != nil {
		panic(v)
	}
}

// Must is a generic helper that wraps a call to a function returning (T, error)
// and panics if the error is non-nil. It is intended for use in variable
// initializations.
func Must[T any](v T, err error) T {
	Nil(err)
	return v
}
