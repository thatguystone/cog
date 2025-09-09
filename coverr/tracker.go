// Package coverr helps you achieve coverage of error branches throughout your
// code base
package coverr

import (
	"crypto/sha256"
	"hash"
	"strconv"
	"sync"

	"github.com/thatguystone/cog/assert"
	"github.com/thatguystone/cog/callstack"
)

// Tracker keeps track of which stack traces it has seen and only returns an
// error for new ones. When used with mocked interfaces that return errors, it
// can be used to exhaust all error paths by repeatedly calling into the test
// target until a nil error is returned.
type Tracker struct {
	mtx    sync.Mutex
	stacks map[[sha256.Size]byte]struct{}
}

// Err returns a generic error if this should fail, or nil
func (trk *Tracker) Err() error {
	h := hasherPool.Get().(*hasher)
	defer hasherPool.Put(h)

	h.h.Reset()

	for f := range callstack.GetSkip(1).Frames() {
		// Avoid allocations for speed
		buf := strconv.AppendUint(h.buf, uint64(f.PC()), 10)
		buf = append(buf, '-')
		buf = append(buf, f.File()...)
		buf = append(buf, ':')
		buf = strconv.AppendInt(buf, int64(f.Line()), 10)
		buf = append(buf, '\n')

		_, err := h.h.Write(buf)
		h.buf = buf[:0]
		assert.Nil(err)
	}

	h.h.Sum(h.sum[:0])

	trk.mtx.Lock()
	defer trk.mtx.Unlock()

	if trk.stacks == nil {
		trk.stacks = map[[sha256.Size]byte]struct{}{}
	}

	// TODO(as): use a sets.Set here?
	if _, ok := trk.stacks[h.sum]; !ok {
		trk.stacks[h.sum] = struct{}{}
		return Err
	}

	return nil
}

type trackerError struct{}

// Err is the value returned by [Tracker.Err]. It can be used with [errors.Is]
// to ensure that actual errors aren't missed.
var Err error = trackerError{}

// Error implements [error.Error]
func (err trackerError) Error() string {
	return "forced to fail by coverr.Tracker"
}

type hasher struct {
	buf []byte
	h   hash.Hash
	sum [sha256.Size]byte
}

var hasherPool = sync.Pool{
	New: func() any {
		return &hasher{
			buf: make([]byte, 0, 256),
			h:   sha256.New(),
		}
	},
}
