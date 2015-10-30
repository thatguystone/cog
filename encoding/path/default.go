package path

import (
	"bufio"
	"io"
)

var defSep = NewSeparator('/')

// Marshal turns the given struct into a structured path
func Marshal(p interface{}) (b []byte, err error) {
	return defSep.Marshal(p)
}

// MarshalInto works exactly like Marshal, except it writes the path to the
// given Writer instead of returning a [ ]byte.
func MarshalInto(p interface{}, w io.Writer) error {
	return defSep.MarshalInto(p, w)
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func Unmarshal(b []byte, p interface{}) error {
	return defSep.Unmarshal(b, p)
}

// UnmarshalReader works exactly like Unmarshal, except it reads from the given
// reader instead of a [ ]byte. Be warned: a bufio.Reader is used, so this might
// over-read.
func UnmarshalReader(r io.Reader, p interface{}) error {
	return defSep.UnmarshalReader(r, p)
}

// UnmarshalBufio works exactly like Unmarshal, except it reads from the given
// Reader.
func UnmarshalBufio(r *bufio.Reader, p interface{}) error {
	return defSep.UnmarshalBufio(r, p)
}
