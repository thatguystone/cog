package integrationtest

//go:generate go run $GOPATH/src/github.com/thatguystone/cog/cmd/cog-path/main.go

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

	Hand Handmade
}

type Ptrs struct {
	A *int32
	B *uint32
	C *string
	D **string
	E ***string

	F *struct {
		G string
		H *bool
	}

	J [2]struct {
		K string
		L bool
	}

	M [2]*struct {
		N string
		O bool
	}

	Hand  *Handmade
	Handy **Handmade
}
