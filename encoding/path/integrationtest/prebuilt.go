package integrationtest

import "github.com/thatguystone/cog/encoding/path"

type Handmade struct{}

func (h Handmade) MarshalPath(e path.Encoder) path.Encoder {
	return e
}

func (h *Handmade) UnmarshalPath(d path.Decoder) path.Decoder {
	return d
}
