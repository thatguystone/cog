package clog

import "time"

// Data is a collection of fields to add to a log entry
type Data map[string]interface{}

// Entry is one complete log entry
type Entry struct {
	// When this Entry was originally logged
	Time time.Time `json:"time"`

	// Which module this Entry belongs to
	Module string `json:"module"`

	// Level of this Entry
	Level Level `json:"level"`

	// Source file and line from where this was logged
	Src string `json:"src"`

	// How much of the call stack to ignore when looking for a file:lineno
	Depth int `json:"-"`

	// Formatted text
	Msg string `json:"msg"`

	// Data to include with the Entry
	Data Data `json:"data"`

	// Ignore logging any errors that occur while logging. This is mainly to
	// avoid logging errors with logging errors with logging errors (aka.
	// infinite recursion while logging errors about logging).
	ignoreErrors bool
}
