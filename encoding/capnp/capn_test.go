package capnp

import (
	"testing"

	// TODO(astone): upgrade to github.com/zombiezen/go-capnproto2
	capnproto "github.com/glycerine/go-capnproto"
	"github.com/iheartradio/cog/check"
)

func TestProto(t *testing.T) {
	check.New(t)

	seg := capnproto.NewBuffer(nil)
	seg.NewPointerList(8)

	FromBytes(ToBytes(seg))
}
