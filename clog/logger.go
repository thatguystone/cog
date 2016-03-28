package clog

// A Logger provides a simplified Log
type Logger interface {
	// With merges the given data into any current data and returns a new
	// Logger with the resulting data. New values replace older ones with the
	// same key.
	With(d Data) Logger

	// Add the KV pair and get a new Logger.
	WithKV(k string, v interface{}) Logger

	// Debug entries: things that can be ignored in production
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	// Info: stuff that is useful, even in production
	Info(args ...interface{})
	Infof(format string, args ...interface{})

	// Something might start going wrong soon
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	// Something went wrong
	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	// A horrible error happened.
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	// Abort the program. There's no recovering from this.
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
}
