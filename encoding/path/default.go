package path

var defSep = Separator('/')

// Marshal turns the given struct into a structured path
func Marshal(p interface{}, cache []byte) (b []byte, err error) {
	return defSep.Marshal(p, cache)
}

// MustMarshal is like Marshal, except it panics on failure
func MustMarshal(p interface{}, cache []byte) []byte {
	return defSep.MustMarshal(p, cache)
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func Unmarshal(b []byte, p interface{}) (unused []byte, err error) {
	return defSep.Unmarshal(b, p)
}

// MustUnmarshal is like Unmarshal, except it panics on failure
func MustUnmarshal(b []byte, p interface{}) (unused []byte) {
	return defSep.MustUnmarshal(b, p)
}
