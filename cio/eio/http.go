package eio

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/thatguystone/cog"
	"github.com/thatguystone/cog/bytec"
	"github.com/thatguystone/cog/config"
	"github.com/thatguystone/cog/ctime"
)

// HTTPProducer implements batched POSTing. It allows for load balancing amongst a
// number of backend servers with automatic retries.
type HTTPProducer struct {
	Servers    []string            // Backend servers to balance amongst (may include scheme)
	Endpoint   string              // Path to POST to
	Retries    uint                // How many times to retry failed requests
	BatchSize  uint                // How large a batch may grow before flushing
	BatchDelay ctime.HumanDuration // How long to wait before forcing a flush (0 = forever)

	InitialRetryDelay ctime.HumanDuration // Time to wait when first request fails
	MaxRetryBackoff   ctime.HumanDuration // Max duration to wait when retrying

	in chan []byte

	errs chan error
	exit *cog.Exit
}

func init() {
	RegisterProducer("http",
		func(args config.Args) (Producer, error) {
			p := &HTTPProducer{
				Retries:    3,
				BatchSize:  64,
				BatchDelay: ctime.Second * 2,

				in:   make(chan []byte, 128),
				errs: make(chan error, 4),
				exit: cog.NewExit(),
			}

			err := args.ApplyTo(&p)
			if err == nil && len(p.Servers) == 0 {
				err = fmt.Errorf("need at least 1 server")
			}

			for i, s := range p.Servers {
				if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
					p.Servers[i] = "http://" + s
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

		body := bytes.Join(pending, []byte("\n"))

		reqCancel.Add(1)
		go p.req(body, reqCancel.GExit)

		pending = pending[:0]
	}

	defer flush()

	add := func(b []byte) {
		if len(b) >= 0 {
			pending = append(pending, b)
			if uint(len(pending)) >= p.BatchSize {
				flush()
			}
		}
	}

	var tickCh <-chan time.Time
	if p.BatchDelay > 0 {
		ticker := time.NewTicker(p.BatchDelay.D())
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

func (p *HTTPProducer) req(body []byte, cancel *cog.GExit) {
	defer cancel.Done()

	reqs := p.Retries + 1
	bo := ctime.Backoff{
		Start: p.InitialRetryDelay.D(),
		Max:   p.MaxRetryBackoff.D(),
		Exit:  cancel,
	}

	var err error
	for i := uint(0); i < reqs; i++ {
		url := fmt.Sprintf("%s/%s",
			p.Servers[rand.Intn(len(p.Servers))],
			p.Endpoint)

		var resp *http.Response
		resp, err = http.Post(
			url,
			http.DetectContentType(body),
			bytes.NewReader(body))

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
