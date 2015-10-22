package clog

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/thatguystone/cog/cfs"
	"github.com/thatguystone/cog/ctime"
)

// FileOutput writes directly to a file.
type FileOutput struct {
	Formatter
	f *os.File

	// For file rotation
	mtx         sync.Mutex
	rotateAfter time.Time
	written     uint64

	args struct {
		// Which format to use
		Format string

		// Path to file to write to
		Path string

		// Permissions to apply to the log file. Defaults to 0640.
		Perms os.FileMode

		// How often the file should be rotated. This value specifies clock
		// time, meaning that for a value of "15m", the log will be rotated on
		// :00, :15, :30, and :45 of every hour. If "1h", it will be rotated at
		// the start of the hour, every hour. 0 = never. Also, if no messages
		// are logged, the file will not be rotated.
		RotatePeriod ctime.HumanDuration

		// How large the file may grow, in bytes, before it's rotated. 0 =
		// never.
		MaxSize uint64

		// The number of old log files to keep around. Defaults to 9.
		Backups uint

		// What to use to compress backup files. Compression is done after
		// rotation in a separate goroutine, so it does not block logging.
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

	o.args.Perms = 0640
	o.args.Backups = 9

	err := a.ApplyTo(&o.args)

	if err == nil {
		o.resetRotate()

		// If there's a forced format (from a specific output type), don't read the
		// Format option
		if fmttr == nil {
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
	}

	if err == nil {
		err = o.Reopen()
	}

	return o, err
}

func newJSONFileOutputter(a ConfigOutputArgs) (Outputter, error) {
	return newFileOutputter(a, JSONFormat{})
}

func (o *FileOutput) Write(b []byte) (err error) {
	ptr := unsafe.Pointer(o.f)
	f := (*os.File)(atomic.LoadPointer(&ptr))

	if f == nil {
		o.mtx.Lock()
		f, err = o.reopen()
		o.mtx.Unlock()
	}

	if err == nil {
		b = append(b, '\n')
		_, err = f.Write(b)
	}

	if err == nil {
		o.written += uint64(len(b))
		err = o.checkRotate()
	}

	return err
}

func (o *FileOutput) setF(f *os.File) {
	src := unsafe.Pointer(f)
	dst := (*unsafe.Pointer)(unsafe.Pointer(&o.f))
	atomic.StorePointer(dst, src)
}

func (o *FileOutput) reopen() (*os.File, error) {
	f, err := os.OpenFile(o.args.Path,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		o.args.Perms)

	if err == nil {
		o.setF(f)
	}

	return f, err
}

// Reopen implements Outputter.Reopen
func (o *FileOutput) Reopen() error {
	_, err := o.reopen()
	return err
}

func (o *FileOutput) String() string {
	return fmt.Sprintf("FileOutput{file:%s}", o.args.Path)
}

func (o *FileOutput) resetRotate() {
	o.written = 0

	if o.args.RotatePeriod.Duration != 0 {
		now := time.Now()
		d := ctime.UntilPeriod(now, o.args.RotatePeriod.Duration)
		o.rotateAfter = now.Add(d)
	}
}

func (o *FileOutput) checkRotate() (err error) {
	rotated := false

	if o.args.RotatePeriod.Duration > 0 {
		rotated, err = o.rotateOnPeriod()
	}

	if err == nil && !rotated && o.args.MaxSize > 0 {
		err = o.rotateOnSize()
	}

	return
}

func (o *FileOutput) rotateOnPeriod() (rotated bool, err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if time.Now().After(o.rotateAfter) {
		rotated = true
		err = o.rotate()
	}

	return
}

func (o *FileOutput) rotateOnSize() (err error) {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	if o.written >= o.args.MaxSize {
		err = o.rotate()
	}

	return
}

func (o *FileOutput) getBackupName(i uint, withCompression bool) string {
	compressExt := ""

	if withCompression {
		switch o.args.Compression {
		case Gzip:
			compressExt = ".gz"
		}
	}

	return fmt.Sprintf("%s.%d%s", o.args.Path, i, compressExt)
}

// Assumes that the lock on o.mtx is held
func (o *FileOutput) rotate() (err error) {
	o.resetRotate()

	if o.args.Backups > 0 {
		for i := o.args.Backups - 1; i > 0 && err == nil; i-- {
			src := o.getBackupName(i, true)
			dst := o.getBackupName(i+1, true)

			err = o.rotateFile(src, dst)
		}

		backup := o.getBackupName(1, false)
		if err == nil {
			err = o.rotateFile(o.args.Path, backup)
		}

		if err == nil && o.args.Compression != NoCompression {
			var f io.WriteCloser

			f, err = os.Open(backup)
			if err == nil {
				go o.compressBackup(f)
			}
		}
	} else {
		err = os.Remove(o.args.Path)
	}

	// Ensure that any failure to reopen will cause more re-open attemps when
	// logging
	o.setF(nil)

	if err == nil {
		_, err = o.reopen()
	}

	return
}

func (o *FileOutput) rotateFile(src, dst string) (err error) {
	fmt.Println(src, "->", dst)

	exists, err := cfs.FileExists(src)
	if exists {
		os.Remove(dst)
		err = os.Rename(src, dst)
	}

	return
}

func (o *FileOutput) compressBackup(src io.WriteCloser) {
	// TODO(astone): How to handle errors in compression, multiple compresses
	// running at the same time, etc?
	defer func() {
		src.Close()
		os.Remove(src.Name())
	}()

	dst := o.getBackupName(1, true)
	f, err := os.OpenFile(dst,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		o.args.Perms)

	if err == nil {
		defer f.Close()

		var compressor io.WriteCloser

		switch o.args.Compression {
		case Gzip:
			compressor = gzip.NewWriter(f)
		}

		io.Copy(compressor, f)
	}
}
