package clog

import "github.com/thatguystone/cog/cio/eio"

const (
	defaultConfigFileOutputName = "__default_config_file__"
	defaultTermOutputName       = "__default_term__"
)

// Config specifies basic config for logging.
//
// The Config struct is meant to be embedded directly into some other struct
// that you're Unmarshaling your application's config into (typically, this is a
// struct that is being filled by json.Unmarshal, yaml.Unmarshal, etc on
// application start).
type Config struct {
	// File is the simplest way of configuring this logger. It sets up a
	// JSONFile writing to the given path, with a root logger that only accepts
	// Info and greater.
	File string

	// Identifies all of the places where log entries are written. They keys in
	// this map name the output.
	Outputs map[string]*OutputConfig

	// Identifies all modules that you want to configure. They keys in this map
	// identify the module to work on.
	//
	// If no modules are given, everything at level Info and above goes to the
	// terminal by default.
	Modules map[string]*ModuleConfig
}

// OutputConfig specifies how an output is to be built
type OutputConfig struct {
	// Which Producer to use. This value is case-insensitive.
	//
	// The full list of Producers can be found at:
	// https://godoc.org/github.com/thatguystone/cog/cio/eio
	Prod     string
	ProdArgs eio.Args // Args to pass to the Producer

	// Which Formatter to use for this output.
	Fmt     string
	FmtArgs eio.Args // Args to pass to the Fmt

	// Log level to use for this output. Use Debug to accept all. This is
	// actually an implicit (and required) Filter.
	Level Level

	// Which filters to apply to this output.
	Filters []FilterConfig
}

// ModuleConfig specifies how a module to to be treated
type ModuleConfig struct {
	// The list of outputs to write to. These values come from the keys in the
	// Outputs map in Config.
	Outputs []string

	// Log level to use for this output. Use Debug to accept all. This is
	// actually an implicit (and required) Filter.
	Level Level

	// Which filters to apply to this module
	Filters []FilterConfig

	// By default, messages are propagated until the root logger. If you want
	// messages to stop here, set this to True.
	DontPropagate bool
}

// FilterConfig is for setting up a Filter
type FilterConfig struct {
	// Which the filter to use. This value is case-insensitive.
	Which string

	// Filter arguments
	Args eio.Args
}

var asdasdcfg = Config{
	// If set, this creates a new root module (the module named "" (the empty
	// string)), and it records any message level >= Info to the named file in
	// JSON format.
	File: "/var/log/app.log",

	Outputs: map[string]*OutputConfig{
		// Only errors with level >= Error will be logged here
		"errors": {
			Prod: "file",
			ProdArgs: eio.Args{
				"path": "/var/log/app.error.json.log",
			},
			Fmt:   "json",
			Level: Error,
			Filters: []FilterConfig{
				FilterConfig{
					Which: "exampleFilter",
					Args: eio.Args{
						"exampleConfig": "someValue",
					},
				},
			},
		},

		// All messages will be accepted here
		"debug": {
			Prod: "file",
			ProdArgs: eio.Args{
				"path": "/var/log/app.log.json",
			},
			Fmt:   "human", // Or "logfmt", or any other valid formatter
			Level: Debug,
		},

		// Only errors level >= Warn will be accepted here
		"heroku": {
			Prod: "file",
			ProdArgs: eio.Args{
				"path": "/var/log/app.logfmt",
			},
			Fmt: "human",
			FmtArgs: eio.Args{
				"ShortTime": true,
			},
			Level: Warn,
		},
	},

	Modules: map[string]*ModuleConfig{
		// All messages eventually reach here, unless DontPropagate==true in a
		// module
		"": {
			Outputs: []string{"errors"},
		},

		// This logs all messages level >= Info, where the filter allows the
		// message through, to the debug log. These messages do not propagate to
		// the root.
		"http": {
			Outputs: []string{"debug"},
			Level:   Info,
			Filters: []FilterConfig{
				FilterConfig{
					Which: "exampleFilter",
					Args: eio.Args{
						"exampleConfig": "someValue",
					},
				},
			},
			DontPropagate: true,
		},

		// This logs all messages level >= Warn, to both the heroku and errors
		// outputs. These messages do not propagate to the root.
		"templates": {
			Outputs:       []string{"heroku", "errors"},
			Level:         Warn,
			DontPropagate: true,
		},

		// This logs all messages from the external library to the debug log.
		// These messages also propagate to the root, which will log any error
		// messages. So, effectively, errors from this module will be logged
		// twice.
		"external.library": {
			Outputs: []string{"debug"},
			Level:   Debug,
		},
	},
}
