package clog

import (
	"fmt"
	"os"
	"sync"
)

// FileOutput writes directly to a file.
type FileOutput struct {
	Formatter

	rwmtx sync.RWMutex
	f     *os.File

	args struct {
		// Which format to use
		Format string

		// Path to file to write to
		Path string
	}
}

func init() {
	RegisterOutputter("JSONFile", newJSONFileOutputter)
	RegisterOutputter("File", func(a ConfigArgs) (Outputter, error) {
		return newFileOutputter(a, nil)
	})
}

func newFileOutputter(a ConfigArgs, fmttr Formatter) (Outputter, error) {
	o := &FileOutput{
		Formatter: fmttr,
	}

	err := a.ApplyTo(&o.args)

	// If there's a forced format (from a specific output type), don't read the
	// Format option
	if err == nil && fmttr == nil {
		switch o.args.Format {
		case "", "logfmt":
			fmttr = LogfmtFormat{}

		case "human":
			hf := HumanFormat{}
			err = a.ApplyTo(&hf)
			fmttr = hf

		case "json":
			fmttr = JSONFormat{}

		default:
			err = fmt.Errorf(`unrecognized output format: "%s"`, o.args.Format)
		}

		o.Formatter = fmttr
	}

	if err == nil {
		err = o.Rotate()
	}

	return o, err
}

func newJSONFileOutputter(a ConfigArgs) (Outputter, error) {
	return newFileOutputter(a, JSONFormat{})
}

func (o *FileOutput) Write(b []byte) error {
	b = append(b, '\n')

	o.rwmtx.RLock()
	defer o.rwmtx.RUnlock()

	// File has an internal lock to prevent interleaving. So just need the read
	// lock above to protect access to the FD.
	_, err := o.f.Write(b)

	return err
}

// Rotate implements Outputter.Rotate
func (o *FileOutput) Rotate() error {
	o.rwmtx.Lock()
	defer o.rwmtx.Unlock()

	f, err := os.OpenFile(o.args.Path,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0640)

	if err == nil {
		o.Exit()
		o.f = f
	}

	return err
}

// Exit implements Outputter.Exit
func (o *FileOutput) Exit() {
	if o.f != nil {
		o.f.Close()
	}
}

func (o *FileOutput) String() string {
	return fmt.Sprintf("FileOutput{file:%s}", o.args.Path)
}
