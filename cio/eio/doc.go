// Package eio implements extensible, async io backends. Many of these provide
// batch operations, and you're completely free to add your own.
//
// Producers
//
// The following producers exist. Each one has a corresponding class, with
// documentation below for arguments (in the Args field).
//
// All values are case-insensitive.
//
//     "Blackhole"  Blackhole
//     "File"       FileProducer
//     "HTTP"       HTTPProducer
//     "Stderr"     OutOutput
//     "Stdout"     OutOutput
//     "TestLog"    TestLogProducer
package eio
