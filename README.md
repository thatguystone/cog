# assert [![Build Status](https://travis-ci.org/thatguystone/assert.svg)](https://travis-ci.org/thatguystone/assert) [![GoDoc](https://godoc.org/github.com/thatguystone/assert?status.svg)](https://godoc.org/github.com/thatguystone/assert)

Assert provides dead-simple assertions for golang testing.

```go
func TestExample(t *testing.T) {
	a := assert.A{t}

	// These are just a few of the provided functions. Check out the full
	// documentation for everything.

	a.Equal(1, 1, "the universe is falling apart")
	a.NotEqual(1, 2, "those can't be equal!")

	panics := func() {
		panic("i get nervous sometimes")
	}
	a.Panic(panics, "this should always panic")

	// Get the original *testing.T
	a.T()
}
```
