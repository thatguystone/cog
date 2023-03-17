// Package assert implements assertions for assumed behavior
package assert

// Nil asserts that v must be nil, or it panics
func Nil(v any) {
	if v != nil {
		panic(v)
	}
}
