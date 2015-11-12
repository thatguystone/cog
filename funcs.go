package cog

import "fmt"

// Must ensures that no error occurred, or panics.
func Must(err error, msg string, args ...interface{}) {
	if err != nil {
		panic(fmt.Errorf("%s: %s", fmt.Sprintf(msg, args...), err))
	}
}

// BytesMust ensures that a function that returns a ([]byte, error) does not
// error.
func BytesMust(b []byte, err error) []byte {
	Must(err, "BytesMust failed")
	return b
}
