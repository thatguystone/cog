// Package assert implements assertions for assumed behavior
package assert

// True ensures that the condition is true
func True(v bool) {
	if !v {
		panic("assert error")
	}
}

// Nil asserts that v must be nil, or it panics
func Nil(v any) {
	if v != nil {
		panic(v)
	}
}
