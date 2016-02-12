package cog

import (
	"bytes"
	"errors"
	"fmt"
)

// Errors holds a list of errors that can be combined into a single, line-
// delimited error message
type Errors struct {
	prefix string
	errs   *[]error
}

func (es *Errors) init() {
	if es.errs == nil {
		es.errs = &[]error{}
	}
}

// Empty checks to see if any errors have been added
func (es *Errors) Empty() bool {
	return es.errs == nil || len(*es.errs) == 0
}

// Prefix returns an Errors for which all errors are prefixed with `prefix`.
func (es *Errors) Prefix(prefix string) *Errors {
	es.init()
	return &Errors{
		prefix: es.prefix + prefix,
		errs:   es.errs,
	}
}

// Add adds an error to the collector. It's safe to call this when err==nil.
func (es *Errors) Add(err error) {
	es.Addf(err, "")
}

// Drain removes every error from the given channel until the channel closes
func (es *Errors) Drain(ch <-chan error) {
	for err := range ch {
		es.Add(err)
	}
}

// Addf is like Add, but prefixes the error with the given format and args. If
// err==nil, this does nothing.
func (es *Errors) Addf(err error, format string, args ...interface{}) {
	if err != nil {
		es.init()

		pfxf := "%s"
		if es.prefix != "" {
			pfxf = "%s: "
		}

		if format != "" {
			pfxf += fmt.Sprintf(format+": ", args...)
		}

		err = fmt.Errorf(pfxf+"%v", es.prefix, err)
		*es.errs = append(*es.errs, err)
	}
}

// Error combines all previous errors into a single error, or returns nil if
// there were no errors.
func (es *Errors) Error() error {
	if es.errs == nil || len(*es.errs) == 0 {
		return nil
	}

	b := bytes.Buffer{}

	for _, err := range *es.errs {
		b.WriteString(fmt.Sprintf("%v\n", err))
	}

	return errors.New(b.String())
}
