package path

// DefSep is the path separator that you're typically going to want to use.
const DefSep = Separator('/')

// Marshal turns the given struct into a structured path
func Marshal(v interface{}, cache []byte) (b []byte, err error) {
	return DefSep.Marshal(v, cache)
}

// MustMarshal is like Marshal, except it panics on failure
func MustMarshal(v interface{}, cache []byte) []byte {
	return DefSep.MustMarshal(v, cache)
}

// Unmarshal is the reverse of Marshal, reading a serialized path into a struct.
func Unmarshal(b []byte, v interface{}) (unused []byte, err error) {
	return DefSep.Unmarshal(b, v)
}

// MustUnmarshal is like Unmarshal, except it panics on failure
func MustUnmarshal(b []byte, v interface{}) (unused []byte) {
	return DefSep.MustUnmarshal(b, v)
}
