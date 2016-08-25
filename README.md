## Cog [![Build Status](https://travis-ci.org/iheartradio/cog.svg)](https://travis-ci.org/iheartradio/cog)

Cog is a collection of utils for golang that I tend to use across many of my projects. Rather than building new cogs everywhere, I've just consolidated them all here. Cogs for everyone!

### Modules

Cog consists of the following modules:

| Module        | Docs                                            | Description |
| ------------- | ----------------------------------------------- | ----------- |
| (root)        | [![GoDoc][root-status]][root]                   | generic utils that didn't fit anywhere else
| bytec         | [![GoDoc][bytec-status]][bytec]                 | extra byte slice utils
| cfs           | [![GoDoc][cfs-status]][cfs]                     | filesystem utils
| check         | [![GoDoc][check-status]][check]                 | test assertions and isolated FS utils
| cio           | [![GoDoc][cio-status]][cio]                     | extra io utils
| clog          | [![GoDoc][clog-status]][clog]                   | a logging framework that looks a bit like python's logging
| cnet          | [![GoDoc][cnet-status]][cnet]                   | misc net utils and a socket implementation using channels
| cort          | [![GoDoc][cort-status]][cort]                   | extra sorting utilities
| ctime         | [![GoDoc][ctime-status]][ctime]                 | time utils
| cync          | [![GoDoc][cync-status]][cync]                   | some extra sync utils
| encoding/capn | [![GoDoc][encoding-capn-status]][encoding-capn] | capnproto Marshaling and Unmarshaling
| encoding/path | [![GoDoc][encoding-path-status]][encoding-path] | path Marshaling and Unmarshaling
| node          | [![GoDoc][node-status]][node]                   | get information about the local node
| stack         | [![GoDoc][stack-status]][stack]                 | runtime call stack utils
| statc         | [![GoDoc][statc-status]][statc]                 | application status and stats
| stringc       | [![GoDoc][stringc-status]][stringc]             | extra strings utils
| unsafec       | [![GoDoc][unsafec-status]][unsafec]             | making things more unsafe

[root]: https://godoc.org/github.com/iheartradio/cog
[root-status]: https://godoc.org/github.com/iheartradio/cog?status.svg
[bytec]: https://godoc.org/github.com/iheartradio/cog/bytec
[bytec-status]: https://godoc.org/github.com/iheartradio/cog/bytec?status.svg
[cfs]: https://godoc.org/github.com/iheartradio/cog/cfs
[cfs-status]: https://godoc.org/github.com/iheartradio/cog/cfs?status.svg
[check]: https://godoc.org/github.com/iheartradio/cog/check
[check-status]: https://godoc.org/github.com/iheartradio/cog/check?status.svg
[cio]: https://godoc.org/github.com/iheartradio/cog/cio
[cio-status]: https://godoc.org/github.com/iheartradio/cog/cio?status.svg
[clog]: https://godoc.org/github.com/iheartradio/cog/clog
[clog-status]: https://godoc.org/github.com/iheartradio/cog/clog?status.svg
[cnet]: https://godoc.org/github.com/iheartradio/cog/cnet
[cnet-status]: https://godoc.org/github.com/iheartradio/cog/cnet?status.svg
[cort]: https://godoc.org/github.com/iheartradio/cog/cort
[cort-status]: https://godoc.org/github.com/iheartradio/cog/cort?status.svg
[ctime]: https://godoc.org/github.com/iheartradio/cog/ctime
[ctime-status]: https://godoc.org/github.com/iheartradio/cog/ctime?status.svg
[cync]: https://godoc.org/github.com/iheartradio/cog/cync
[cync-status]: https://godoc.org/github.com/iheartradio/cog/cync?status.svg
[encoding-capn]: https://godoc.org/github.com/iheartradio/cog/encoding/capn
[encoding-capn-status]: https://godoc.org/github.com/iheartradio/cog/encoding/capn?status.svg
[encoding-path]: https://godoc.org/github.com/iheartradio/cog/encoding/path
[encoding-path-status]: https://godoc.org/github.com/iheartradio/cog/encoding/path?status.svg
[node]: https://godoc.org/github.com/iheartradio/cog/node
[node-status]: https://godoc.org/github.com/iheartradio/cog/node?status.svg
[stack]: https://godoc.org/github.com/iheartradio/cog/stack
[stack-status]: https://godoc.org/github.com/iheartradio/cog/stack?status.svg
[statc]: https://godoc.org/github.com/iheartradio/cog/statc
[statc-status]: https://godoc.org/github.com/iheartradio/cog/statc?status.svg
[stringc]: https://godoc.org/github.com/iheartradio/cog/stringc
[stringc-status]: https://godoc.org/github.com/iheartradio/cog/stringc?status.svg
[unsafec]: https://godoc.org/github.com/iheartradio/cog/unsafec
[unsafec-status]: https://godoc.org/github.com/iheartradio/cog/unsafec?status.svg

Each module contains full documentation over on godoc, including tons of examples.

As you might have noticed, the modules have weirdly spelled names; this is so
that you can, for example, import both "sync" and "cync" into the same file,
since "cync" only supplements "sync".
