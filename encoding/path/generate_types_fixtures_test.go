package path

const fixtureBasic = `package test

import "github.com/thatguystone/cog/encoding/path"

type stuff struct {
	A, Z [8]byte
	B    [8]int8
	C    [8]uint8
	D    [8]uint32
	E    [8][2][1]uint32
	F    [8][2]struct {
		Y uint32
		Z bool
	}
	_ path.Static ` + "`path:\"static\"`" + `
	G interfaced
	H BoolInterfaced
	I struct {
		BoolInterfaced
		O uint32
	}
	J uint16
	K embeded

	g uint32
}

type arrayDef [4]uint32
type basicDef int64

type interfaced struct{}

func (interfaced) MarshalPath(s path.Encoder) path.Encoder    { return s }
func (*interfaced) UnmarshalPath(s path.Decoder) path.Decoder { return s }

type BoolInterfaced bool

func (BoolInterfaced) MarshalPath(s path.Encoder) path.Encoder    { return s }
func (*BoolInterfaced) UnmarshalPath(s path.Decoder) path.Decoder { return s }

type embeded struct {
	AA uint32
	BB uint16
}`

const fixtureEndToEnd = `package integrationtest

import "github.com/thatguystone/cog/encoding/path"

type Handmade struct{}
func (h Handmade) MarshalPath(s path.Encoder) path.Encoder { return s }
func (h *Handmade) UnmarshalPath(s path.Decoder) path.Decoder { return s }

type HandmadeSimple bool

func (h HandmadeSimple) MarshalPath(s path.Encoder) path.Encoder {
	return s.EmitBool(bool(h))
}

func (h *HandmadeSimple) UnmarshalPath(s path.Decoder) path.Decoder {
	return s.ExpectBool((*bool)(h))
}

type Deflect struct {
	A uint32
}

type DeflectEmpty struct{}
`

const fixtureIntegrate = `package integrationtest

type Basic struct {
	A int32
	B uint32
	C string
	D struct {
		E bool
		F int8
	}

	G [2]struct {
		H string
		I bool
	}

	J [8][4][2]string

	K [8][4][2]struct {
		L uint32
		M bool
		N [4][2]string
		O [4][2]struct {
			P uint16
		}
	}

	Hand  Handmade
	Hands [4]Handmade
	Handy [2]struct {
		Handmade
		HM Handmade
	}
	Deflect   Deflect
	DeflectE  DeflectEmpty
	Simp      Simple
	HSimp     HandmadeSimple
	SimpArray [2]Simple
	SimpAnon  struct{
		Simple
		B Simple
	}
}

type Embed struct {
	Basic
	A uint16
	Z bool
}

type Redefine Embed

type Mashup struct {
	Embed
	J uint32
}

type Simple int32
`

const fixtureEndToEndTest = `package integrationtest

import (
	"testing"

	"github.com/thatguystone/cog/check"
	"github.com/thatguystone/cog/encoding/path"
)

func TestBasic(t *testing.T) {
	c := check.New(t)

	b := Basic{
		A: -123,
		B: 987,
		C: "basic string",
		Deflect: Deflect{
			A: 999,
		},
		Simp: Simple(123),
		HSimp: HandmadeSimple(true),
		SimpArray: [2]Simple{Simple(1), Simple(2)},
	}

	b.D.E = true
	b.D.F = 9
	b.G[0].H = "h0"
	b.G[0].I = true
	b.G[1].H = "h1"
	b.J[4][2][1] = "buried"
	b.K[2][3][0].L = 9018
	b.K[3][1][0].O[1][0].P = 6182
	b.K[3][0][0].O[2][1].P = 986
	b.SimpAnon.Simple = Simple(913245)
	b.SimpAnon.B = Simple(3245)

	sep := path.Separator('/')
	enc := b.MarshalPath(sep.NewEncoder(nil))
	c.MustNotError(enc.Err)
	c.True(len(enc.B) > 0)

	bout := Basic{}
	dec := bout.UnmarshalPath(sep.NewDecoder(enc.B))
	c.MustNotError(dec.Err)

	c.Equal(b, bout)
}

func TestMashup(t *testing.T) {
	c := check.New(t)

	m := Mashup{
		Embed: Embed{
			Basic: Basic{
				A: 123,
			},
			Z: true,
		},
	}

	sep := path.Separator('/')
	enc := m.MarshalPath(sep.NewEncoder(nil))
	c.MustNotError(enc.Err)
	c.True(len(enc.B) > 0)

	mout := Mashup{}
	dec := mout.UnmarshalPath(sep.NewDecoder(enc.B))
	c.MustNotError(dec.Err)

	c.Equal(m, mout)
}
`
