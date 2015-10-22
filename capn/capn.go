// Package capn implements some stupid stuff for capnproto
package capn

import (
	"bytes"

	capnproto "github.com/glycerine/go-capnproto"
	"github.com/thatguystone/cog"
)

// TODO(astone): upgrade to github.com/zombiezen/go-capnproto2

// ProtoFromBytes loads a segment from a byte slice
func ProtoFromBytes(b []byte) *capnproto.Segment {
	seg, _, err := capnproto.ReadFromMemoryZeroCopy(b)
	cog.Must(err, "failed to decode proto")

	return seg
}

// ProtoToBytes serializes a segment to a byte slice
func ProtoToBytes(seg *capnproto.Segment) []byte {
	b := bytes.Buffer{}
	_, err := seg.WriteTo(&b)
	cog.Must(err, "failed to write to buffer")

	return b.Bytes()
}
