package clog

import (
	"fmt"
	"os"
	"sync"

	"github.com/thatguystone/cog/config"
)

// FileOutput writes directly to a file.
//
// If you're using the "human" log formatter, you may also include its arguments
// in the file's arguments.
//
// Examples:
//
//  JSON:
//     Config{
//         Outputs: map[string]*OutputConfig{
//             "human": {
//                 Which: "JSONFile",
//                 Level: clog.Debug,
//                 Args: config.Args{
//                     "Path": "/var/log/file.json.log",
//                 },
//             },
//         },
//
//  Human:
//     Config{
//         Outputs: map[string]*OutputConfig{
//             "human": {
//                 Which: "File",
//                 Level: clog.Debug,
//                 Format: FormatterConfig{
//                     Name: "human", // Or "logfmt", or any other valid formatter
//                 },
//                 Args: config.Args{
//                     "Path": "/var/log/file.log",
//                 },
//             },
//         },
//     }
type FileOutput struct {
	Formatter

	rwmtx sync.RWMutex
	f     *os.File

	Args struct {
		// Path to file to write to
		Path string
	}
}

func init() {
	RegisterOutputter("JSONFile",
		FormatterConfig{Name: "JSON"},
		newFileOutputter)
	RegisterOutputter("File",
		FormatterConfig{Name: "logfmt"},
		newFileOutputter)
}

func newFileOutputter(a config.Args, f Formatter) (Outputter, error) {
	o := &FileOutput{
		Formatter: f,
	}

	err := a.ApplyTo(&o.Args)
	if err == nil {
		err = o.Rotate()
	}

	return o, err
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

	f, err := os.OpenFile(o.Args.Path,
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
	return fmt.Sprintf("FileOutput{file:%s}", o.Args.Path)
}
