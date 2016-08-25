package eio

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/iheartradio/cog"
	"github.com/iheartradio/cog/bytec"
	"github.com/iheartradio/cog/ctime"
)

// HTTPProducer implements batched POSTing. It allows for load balancing amongst a
// number of backend servers with automatic retries.
type HTTPProducer struct {
	Args struct {
		Servers    []string            // Backend servers to balance amongst (may include scheme)
		Endpoint   string              // Path to POST to
		Retries    uint                // How many times to retry failed requests
		BatchSize  uint                // How large a batch may grow before flushing
		BatchDelay ctime.HumanDuration // How long to wait before forcing a flush (0 = forever)

		InitialRetryDelay ctime.HumanDuration // Time to wait when first request fails
		MaxRetryBackoff   ctime.HumanDuration // Max duration to wait when retrying

		// Some servers don't know how to handle Chunked bodies.
		DisableChunked bool

		// MimeType to use instead of "application/octet-stream".
		MimeType string
	}

	in chan []byte

	errs chan error
	exit *cog.Exit
}

const httpContentType = "application/octet-stream"

var httpNewline = []byte("\n")

func init() {
	RegisterProducer("http",
		func(args Args) (Producer, error) {
			p := &HTTPProducer{
				in:   make(chan []byte, 128),
				errs: make(chan error, 4),
				exit: cog.NewExit(),
			}

			p.Args.Retries = 3
			p.Args.BatchSize = 64
			p.Args.BatchDelay = ctime.Second * 2

			err := args.ApplyTo(&p.Args)
			if err == nil && len(p.Args.Servers) == 0 {
				err = fmt.Errorf("need at least 1 server")
			}

			// A newline is added after every message. This fixes batch size.
			p.Args.BatchSize *= 2

			if p.Args.MimeType == "" {
				p.Args.MimeType = httpContentType
			}

			for i, s := range p.Args.Servers {
				if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
					p.Args.Servers[i] = "http://" + s
				}
			}

			if err == nil {
				p.exit.Add(1)
				go p.run()
			}

			return p, err
		})
}

func (p *HTTPProducer) run() {
	reqCancel := cog.NewExit()
	defer func() {
		reqCancel.Exit()
		close(p.errs)
		p.exit.Done()
	}()

	var pending [][]byte

	flush := func() {
		if len(pending) == 0 {
			return
		}

		var r io.Reader
		if p.Args.DisableChunked {
			r = bytes.NewReader(bytes.Join(pending, nil))
		} else {
			r = bytec.MultiReader(pending...)
		}

		reqCancel.Add(1)
		go p.req(r, reqCancel.GExit)

		pending = pending[:0]
	}

	defer flush()

	add := func(b []byte) {
		if len(b) >= 0 {
			pending = append(pending, b, httpNewline)
			if uint(len(pending)) >= p.Args.BatchSize {
				flush()
			}
		}
	}

	var tickCh <-chan time.Time
	if p.Args.BatchDelay > 0 {
		ticker := time.NewTicker(p.Args.BatchDelay.D())
		defer ticker.Stop()
		tickCh = ticker.C
	}

	for {
		select {
		case b := <-p.in:
			add(b)

		case <-tickCh:
			flush()

		case <-p.exit.C:
			return
		}
	}
}

func (p *HTTPProducer) req(body io.Reader, cancel *cog.GExit) {
	defer cancel.Done()

	reqs := p.Args.Retries + 1
	bo := ctime.Backoff{
		Start: p.Args.InitialRetryDelay.D(),
		Max:   p.Args.MaxRetryBackoff.D(),
		Exit:  cancel,
	}

	var err error
	for i := uint(0); i < reqs; i++ {
		url := fmt.Sprintf("%s/%s",
			p.Args.Servers[rand.Intn(len(p.Args.Servers))],
			p.Args.Endpoint)

		var resp *http.Response
		resp, err = http.Post(url, p.Args.MimeType, body)

		if err == nil {
			resp.Body.Close()
			if (resp.StatusCode / 100) != 2 { // Any 2** code is OK
				err = fmt.Errorf("got status %d", resp.StatusCode)
			}
		}

		if err == nil {
			return
		}

		if !bo.Wait() {
			break
		}
	}

	if err != nil {
		p.errs <- err
	}
}

// Errs implements Producer.Errs
func (p *HTTPProducer) Errs() <-chan error { return p.errs }

// Rotate implements Producer.Rotate
func (p *HTTPProducer) Rotate() error { return nil }

// Produce implements Producer.Produce
func (p *HTTPProducer) Produce(b []byte) {
	select {
	case p.in <- bytec.Dup(b):
	case <-p.exit.C:
	}
}

// Close implements Producer.Close
func (p *HTTPProducer) Close() (es cog.Errors) {
	p.exit.Signal()
	es.Drain(p.errs)
	p.exit.Exit()

	return
}
