// Package clog implements a python-like, module-based logger with a variety of
// backends and formats.
//
// This logger is based on Python's fantastic logging package, though it intends
// to be far simpler. Every part of the system should have its own individual
// logger, specific by a module name, so that its messages can be easily
// distinguished from other parts. In golang's log package, people have
// typically resorted to prefixing messages with things like "[INFO module]",
// but that is tedious, error prone, rather un-enforceable for external
// pacakges, and impossible to serialize for all logging services.
//
// Clog takes the approach that data should be represented as data, not strings
// (there are methods such as Infod, Warnd, etc), and this data is properly
// serialized for all output formats. Each message includes a timestamp, the
// module that generated the message, the level, and the source file and line.
//
// Like python's logging, clog modules are heirachical, in that rules for
// something like "server.http" will be applied to "server.http.requests" and
// "server.http.responses" as the messages propogate. Each module may have
// multiple outputs, so you may log to both a file and logstash for the same
// module, if you wish.
//
// Configuration
//
// Configuring logging packages has always been a rather daunting affair, but
// clog takes great pains to be as simple as possible.
//
// First, there are 2 concepts in logging: Outputs and Modules. Outputs are
// where log messages are written to, ie. a file, an external service, syslog,
// etc. Modules are where messages come from, ie. "http" for the http interface,
// "library.name" for some library named "name". You may choose your own values
// for these, just know that they're arranged in a tree, such that
// "http.request" is a child of "http" and messages from it propagate to "http".
//
// Since examples are worth more than descriptions, let's take a look at a
// pretty complex configuration, with comments explaining how it all works
// together (this is much simpler to write in JSON, but you get the point).
//
//     Config{
//         // If set, this creates a new root module (the module named "" (the empty
//         // string)), and it records any message level >= Info to the named file in
//         // JSON format.
//         File: "/var/log/app.log",
//
//         Outputs: map[string]*OutputConfig{
//             // Only errors with level >= Error will be logged here
//             "errors": {
//                 Prod: "file",
//                 ProdArgs: eio.Args{
//                     "path": "/var/log/app.error.json.log",
//                 },
//                 Fmt:   "json",
//                 Level: Error,
//                 Filters: []FilterConfig{
//                     FilterConfig{
//                         Which: "exampleFilter",
//                         Args: eio.Args{
//                             "exampleConfig": "someValue",
//                         },
//                     },
//                 },
//             },
//
//             // All messages will be accepted here
//             "debug": {
//                 Prod: "file",
//                 ProdArgs: eio.Args{
//                     "path": "/var/log/app.log.json",
//                 },
//                 Fmt:   "human", // Or "logfmt", or any other valid formatter
//                 Level: Debug,
//             },
//
//             // Only errors level >= Warn will be accepted here
//             "heroku": {
//                 Prod: "file",
//                 ProdArgs: eio.Args{
//                     "path": "/var/log/app.logfmt",
//                 },
//                 Fmt: "logfmt",
//                 FmtArgs: eio.Args{
//                     "ShortTime": true,
//                 },
//                 Level: Warn,
//             },
//         },
//
//         Modules: map[string]*ModuleConfig{
//             // All messages eventually reach here, unless DontPropagate==true in a
//             // module
//             "": {
//                 Outputs: []string{"errors"},
//             },
//
//             // This logs all messages level >= Info, where the filter allows the
//             // message through, to the debug log. These messages do not propagate to
//             // the root.
//             "http": {
//                 Outputs: []string{"debug"},
//                 Level:   Info,
//                 Filters: []FilterConfig{
//                     FilterConfig{
//                         Which: "exampleFilter",
//                         Args: eio.Args{
//                             "exampleConfig": "someValue",
//                         },
//                     },
//                 },
//                 DontPropagate: true,
//             },
//
//             // This logs all messages level >= Warn, to both the heroku and errors
//             // outputs. These messages do not propagate to the root.
//             "templates": {
//                 Outputs:       []string{"heroku", "errors"},
//                 Level:         Warn,
//                 DontPropagate: true,
//             },
//
//             // This logs all messages from the external library to the debug log.
//             // These messages also propagate to the root, which will log any error
//             // messages. So, effectively, errors from this module will be logged
//             // twice.
//             "external.library": {
//                 Outputs: []string{"debug"},
//                 Level:   Debug,
//             },
//         },
//     }
//
// Producers
//
// A full list of Producers can be found at:
// https://godoc.org/github.com/iheartradio/cog/cio/eio
//
// Filters
//
// The following filters are available (their names are case-insensitive):
//
//     "Level"  LevelFilter
//
// Formatters
//
// The following formatters exist; their arguments are specified in the listed
// classes. Like everything else, these names are case-insensitive.
//
//     "Human"   HumanFormat
//     "JSON"    JSONFormat
//     "LogFmt"  LogFmtFormat
//
// Testing
//
// If you want log output to go to a test's log, check out:
// https://godoc.org/github.com/iheartradio/cog/check/chlog
package clog
