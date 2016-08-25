package eio

import (
	"bytes"
	"os"
	"sync"

	"github.com/iheartradio/cog"
)

// FileProducer writes each message as a single line to the given file
type FileProducer struct {
	rwmtx sync.RWMutex
	once  sync.Once
	f     *os.File
	errs  chan error

	Args struct {
		// Path of file to write to
		Path string
	}
}

func init() {
	RegisterProducer("file",
		func(args Args) (Producer, error) {
			p := &FileProducer{
				errs: make(chan error, 8),
			}

			err := args.ApplyTo(&p.Args)
			if err == nil {
				err = p.Rotate()
			}

			return p, err
		})
}

// Produce implements Producer.Produce
func (p *FileProducer) Produce(b []byte) {
	b = append(bytes.TrimSpace(b), '\n')

	p.rwmtx.RLock()
	defer p.rwmtx.RUnlock()

	// File has an internal lock to prevent interleaving. So just need the read
	// lock above to protect access to the FD.
	_, err := p.f.Write(b)

	if err != nil {
		p.errs <- err
	}
}

// Errs implements Producer.Errs
func (p *FileProducer) Errs() <-chan error {
	return p.errs
}

// Rotate implements Producer.Rotate
func (p *FileProducer) Rotate() error {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	f, err := os.OpenFile(p.Args.Path,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0640)

	if err == nil {
		p.close()
		p.f = f
	}

	return err
}

func (p *FileProducer) close() {
	if p.f != nil {
		p.f.Close()
	}
}

// Close implements Producer.Close
func (p *FileProducer) Close() (es cog.Errors) {
	p.once.Do(func() {
		p.close()
		close(p.errs)

		es.Drain(p.errs)
	})
	return
}
