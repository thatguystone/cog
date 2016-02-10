// Package unsafec makes unsafe even more unsafe.
package unsafec

import "unsafe"

// String turns a byte slice into a string. This is unsafe as it leaves
// a string mutable. Don't use this if you will EVER modify the underlying
// byte slice.
//
// From: https://github.com/golang/go/issues/2632#issuecomment-66061057
//
// As bradfitz says, "As a medium-term hack, I showed they could do... But I
// felt bad even pointing that out."
//
// Until golang can optimize this case, sadness everywhere.
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Bytes is the opposite String. This is unsafe as it leaves a string mutable.
// Don't not, under any circumstances, modify the returned byte slice.
// Seriously, if you do, unicorns will die, kittens will commit suicide, and
// puppies will maul you.
func Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
