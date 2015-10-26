// Package bytec implements some bytes utilities.
//
// "bytec" is pronounced "bytes".
package bytec

// Dup makes a copy of the given byte slice
func Dup(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
