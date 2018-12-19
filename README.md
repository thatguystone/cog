## Cog [![Build Status](https://travis-ci.org/thatguystone/cog.svg)](https://travis-ci.org/thatguystone/cog)

Cog is a collection of utils for golang that I tend to use across many of my
projects. Rather than building new cogs everywhere, I've just consolidated
them all here. Cogs for everyone!

### Modules

Cog consists of the following modules:

| Module        | Docs                                 | Description |
| ------------- | ------------------------------------ | ----------- |
| cfs           | [![GoDoc][cfs-status]][cfs]          | filesystem utils
| check         | [![GoDoc][check-status]][check]      | test assertions
| cio           | [![GoDoc][cio-status]][cio]          | extra io utils
| ctime         | [![GoDoc][ctime-status]][ctime]      | time utils
| stringc       | [![GoDoc][stringc-status]][stringc]  | extra strings utils

[cfs]: https://godoc.org/github.com/thatguystone/cog/cfs
[cfs-status]: https://godoc.org/github.com/thatguystone/cog/cfs?status.svg
[check]: https://godoc.org/github.com/thatguystone/cog/check
[check-status]: https://godoc.org/github.com/thatguystone/cog/check?status.svg
[cio]: https://godoc.org/github.com/thatguystone/cog/cio
[cio-status]: https://godoc.org/github.com/thatguystone/cog/cio?status.svg
[ctime]: https://godoc.org/github.com/thatguystone/cog/ctime
[ctime-status]: https://godoc.org/github.com/thatguystone/cog/ctime?status.svg
[stringc]: https://godoc.org/github.com/thatguystone/cog/stringc
[stringc-status]: https://godoc.org/github.com/thatguystone/cog/stringc?status.svg

Each module contains full documentation over on godoc, including tons of examples.

As you might have noticed, the modules have weirdly spelled names; this is so
that you can, for example, import both "strings" and "stringc" into the same
file, since "stringc" only supplements "strings".
