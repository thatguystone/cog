package clog

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"
)

// FileOutput writes directly to a file.
type FileOutput struct {
	Formatter
	f *os.File

	// For file rotation
	mtx sync.Mutex

	args struct {
		// Which format to use
		Format string

		// Path to file to write to
		Path string
	}
}

func init() {
	RegisterOutputter("JSONFile", newJSONFileOutputter)
	RegisterOutputter("File", func(a ConfigOutputArgs) (Outputter, error) {
		return newFileOutputter(a, nil)
	})
}

func newFileOutputter(a ConfigOutputArgs, fmttr Formatter) (Outputter, error) {
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
			fmttr = HumanFormat{}

		case "json":
			fmttr = JSONFormat{}

		default:
			err = fmt.Errorf(`unrecognized output format: "%s"`, o.args.Format)
		}

		o.Formatter = fmttr
	}

	if err == nil {
		err = o.Reopen()
	}

	return o, err
}

func newJSONFileOutputter(a ConfigOutputArgs) (Outputter, error) {
	return newFileOutputter(a, JSONFormat{})
}

func (o *FileOutput) Write(b []byte) error {
	ptr := unsafe.Pointer(o.f)
	f := (*os.File)(atomic.LoadPointer(&ptr))

	b = append(b, '\n')
	_, err := f.Write(b)

	return err
}

// Reopen implements Outputter.Reopen
func (o *FileOutput) Reopen() error {
	f, err := os.OpenFile(o.args.Path,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0640)

	if err == nil {
		src := unsafe.Pointer(f)
		dst := (*unsafe.Pointer)(unsafe.Pointer(&o.f))
		atomic.StorePointer(dst, src)
	}

	return err
}

func (o *FileOutput) String() string {
	return fmt.Sprintf("FileOutput{file:%s}", o.args.Path)
}
