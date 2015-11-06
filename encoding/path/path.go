// Package path provides Marshaling and Unmarshaling for [ ]byte-encoded paths.
//
// For example, if you have a path like "/pies/apple/0/", of the standard form
// "/pies/<Type:string>/<Index:uint32>/", this helps you Unmarshal that into a
// Pies struct with the fields "Type" and "Index", and vis-versa.
//
// This encodes into a binary, non-human-readable format when using anything but
// strings. It also only supports Marshaling and Unmarshaling into structs using
// primitive types (with fixed byte sizes), which means that your paths must be
// strictly defined before using this.
//
// Warning: do not use this for anything that will change over time and that
// needs to remain backwards compatible. For that, you should use something like
// capnproto.
package path

import "reflect"

// Static is used to place a static, unchanging element into the path.
type Static struct{}

// Separator allows you to change the path separator used.
type Separator byte

// State is the state of the Marshaler or Unmarshaler.
type State struct {
	// Where bytes are appended to and removed from
	B []byte

	// Any error encountered along the way
	Err error

	s byte
}

var staticType = reflect.TypeOf(Static{})
