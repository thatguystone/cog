package bytec

import "io"

type multiReader struct {
	rs [][]byte
}

// MultiReader returns a Reader that's the logical concatenation of the
// provided input byte slice.
func MultiReader(bs ...[]byte) io.Reader {
	rs := make([][]byte, len(bs))
	copy(rs, bs)
	return &multiReader{rs: rs}
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	for len(mr.rs) > 0 {
		r := mr.rs[0]
		n = copy(p, r)

		r = r[n:]
		if len(r) == 0 {
			mr.rs = mr.rs[1:]
		} else {
			mr.rs[0] = r
		}

		if n > 0 {
			return
		}
	}

	return 0, io.EOF
}
