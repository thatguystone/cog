package path

import "bytes"

var defSep = NewSeparator('/')

// Marshal turns the given struct into a structured path
func Marshal(p interface{}) (b []byte, err error) {
	return defSep.Marshal(p)
}

// MarshalInto works exactly like Marshal, except it writes the path to the
// given Buffer instead of returning a [ ]byte.
func MarshalInto(p interface{}, buff *bytes.Buffer) error {
	return defSep.MarshalInto(p, buff)
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func Unmarshal(b []byte, p interface{}) error {
	return defSep.Unmarshal(b, p)
}

// UnmarshalFrom works exactly like Unmarshal, except it reads from the given
// Buffer instead of a [ ]byte.
func UnmarshalFrom(buff *bytes.Buffer, p interface{}) error {
	return defSep.UnmarshalFrom(buff, p)
}
