// Package capnp implements some stupid stuff for capnproto
package capnp

import (
	"bytes"

	capnproto "github.com/glycerine/go-capnproto"
	"github.com/iheartradio/cog"
)

// TODO(astone): upgrade to github.com/zombiezen/go-capnproto2

// FromBytes loads a segment from a byte slice
func FromBytes(b []byte) *capnproto.Segment {
	seg, _, err := capnproto.ReadFromMemoryZeroCopy(b)
	cog.Must(err, "failed to decode proto")

	return seg
}

// ToBytes serializes a segment to a byte slice
func ToBytes(seg *capnproto.Segment) []byte {
	b := bytes.Buffer{}
	_, err := seg.WriteTo(&b)
	cog.Must(err, "failed to write to buffer")

	return b.Bytes()
}
