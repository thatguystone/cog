package clog

import "encoding/json"

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
	Outputs map[string]*ConfigOutput

	// Identifies all modules that you want to configure. They keys in this map
	// identify the module to work on.
	//
	// If no modules are given, everything at level Info and above goes to the
	// terminal by default.
	Modules map[string]*ConfigModule
}

// ConfigOutput specifies how an output is to be built
type ConfigOutput struct {
	// Which Outputter to use. This value is case-insensitive.
	Which string

	// Log level to use for this output. Use Debug to accept all. This is
	// actually an implicit (and required) Filter.
	Level Level

	// Which filters to apply to this output.
	Filters []ConfigFilter

	// Arguments to provide to the underlying Outputter (the one specified by
	// Which above).
	Args ConfigArgs
}

// ConfigModule specifies how a module to to be treated
type ConfigModule struct {
	// The list of outputs to write to. These values come from the keys in the
	// Outputs map in Config.
	Outputs []string

	// Log level to use for this output. Use Debug to accept all. This is
	// actually an implicit (and required) Filter.
	Level Level

	// Which filters to apply to this module
	Filters []ConfigFilter

	// By default, messages are propagated until the root logger. If you want
	// messages to stop here, set this to True.
	DontPropagate bool
}

// ConfigFilter is for setting up a filter
type ConfigFilter struct {
	// Which the filter to use. This value is case-insensitive.
	Which string

	// Filter arguments
	Args ConfigArgs
}

// ConfigArgs is passed directly to an Outputter or Filter when being created.
// See the individual Outputters and Filters for what these arguments may be.
type ConfigArgs map[string]interface{}

// ApplyTo is typically only used by Outputters and Filters when they are
// building themselves. This Unmarshals the options into the given interface
// for simpler configuration.
func (a ConfigArgs) ApplyTo(i interface{}) (err error) {
	if len(a) > 0 {
		var b []byte
		b, err = json.Marshal(a)
		if err == nil {
			err = json.Unmarshal(b, i)
		}
	}

	return
}
