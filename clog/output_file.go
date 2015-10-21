package clog

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/thatguystone/cog"
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

		// How often the file should be rotated. This value specifies clock
		// time, meaning that for a value of "15m", the log will be rotated on
		// :00, :15, :30, and :45 of every hour. If "1h", it will be rotated at
		// the start of the hour, every hour. 0 = never.
		RotatePeriod cog.HumanDuration

		// How large the file may grow, in bytes, before it's rotated. 0 =
		// never.
		MaxSize uint64

		// The number of old log files to keep around. Defaults to 9.
		Backups uint

		// What to use to compress backup files
		Compression FileCompression
	}
}

// FileCompression is the compression that is applied to rotated log files
type FileCompression int

const (
	// NoCompression does not apply any file compression
	NoCompression FileCompression = iota

	// Gzip applies gzip and suffixes rotated files with ".gz"
	Gzip
)

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

	o.args.Backups = 9

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
