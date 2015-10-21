## Cog [![Build Status](https://travis-ci.org/thatguystone/cog.svg)](https://travis-ci.org/thatguystone/cog) [![GoDoc](https://godoc.org/github.com/thatguystone/cog?status.svg)](https://godoc.org/github.com/thatguystone/cog)

Cog is a collection of utilities for golang that I tend to use across many of my projects. Rather than building new cogs everywhere, I've just consolidated them all here. Cogs for everyone!

### Check [![GoDoc](https://godoc.org/github.com/thatguystone/cog/check?status.svg)](https://godoc.org/github.com/thatguystone/cog/check)

Check provides dead-simple assertions for golang testing.

```go
import "github.com/thatguystone/cog/check"

func TestIt(t *testing.T) {
	c := check.New(t)

	// These are just a few of the provided functions. Check out the full
	// documentation for everything.

	c.Equal(1, 1, "the universe is falling apart")
	c.NotEqual(1, 2, "those can't be equal!")

	panics := func() {
		panic("i get nervous sometimes")
	}
	c.Panic(panics, "this should always panic")

	// Get the original *testing.T
	c.T()
}
```
