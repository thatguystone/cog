package statc

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/iheartradio/cog/clog"
)

// HTTPMuxer is a wrapper around httprouter.Router. It intercepts all requests
// and reports useful information about them.
type HTTPMuxer struct {
	// Exposed so that you can set Options on it, use it as the server
	// handler, etc. Any routes added directly here will have no stats
	// recorded for them.
	R    *httprouter.Router
	log  *clog.Log
	name Name
	key  string
	s    *S
}

// An HTTPEndpoint is used for timing requests to an endpoint.
type HTTPEndpoint struct {
	statuses *httpStatuses
}

// An HTTPResp wraps an http.ResponseWriter, providing timing and tracking
// support.
type HTTPResp struct {
	http.ResponseWriter
	statuses *httpStatuses
	start    time.Time
	status   int
}

type httpStatuses struct {
	all    *Timer
	panics *Timer
	m      map[int]*Timer
	s      []*Timer
}

var (
	httpRespPool = sync.Pool{
		New: func() interface{} {
			return new(HTTPResp)
		},
	}

	httpCodes = []int{
		0, // Unknown
		http.StatusContinue,
		http.StatusSwitchingProtocols,
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,
		http.StatusMultipleChoices,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusNotModified,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusProxyAuthRequired,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusTeapot,
		426, // Upgrade
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
	}
)

// NewHTTPMuxer creates a new HTTP muxer that exposes status information for
// all endpoints.
//
// At /_status, information from the last snapshot is accessible with the
// given key (ie. GET "/_status?key=<key>").
func (s *S) NewHTTPMuxer(name string) *HTTPMuxer {
	m := &HTTPMuxer{
		R:    httprouter.New(),
		log:  s.log.Get(name),
		name: s.Name(name),
		key:  s.cfg.StatusKey,
		s:    s,
	}

	m.Handle("GET", "/_status", m.statusHandler)

	return m
}

func (m *HTTPMuxer) statusHandler(
	rw http.ResponseWriter,
	req *http.Request,
	params httprouter.Params) {

	if req.FormValue("key") != m.key {
		http.Error(rw, "", http.StatusForbidden)
		return
	}

	jf := JSONFormat{}
	jf.Args.Pretty = true

	b, err := jf.FormatSnap(m.s.Snapshot())
	if err != nil {
		m.log.Errorf("failed to format snapshot: %v", err)
		http.Error(rw,
			"failed to format snapshot",
			http.StatusInternalServerError)
	} else {
		rw.Write(b)
	}
}

// Endpoint is used to define an HTTP endpoint. If you're not using the
// provided handlers, you can use this to time requests.
func (m *HTTPMuxer) Endpoint(method, path string) HTTPEndpoint {
	path = httprouter.CleanPath(path)
	name := m.name.Join(path, method)

	statuses := m.newHTTPStatuses(name)
	m.s.AddSnapshotter(name, statuses)

	return HTTPEndpoint{
		statuses: statuses,
	}
}

// Handle wraps the given Handle with stats reporting
func (m *HTTPMuxer) Handle(method, path string, handle httprouter.Handle) {
	sRec := m.Endpoint(method, path)

	m.R.Handle(method, path,
		func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			sw := sRec.Start(w)
			defer func() {
				err := recover()
				sw.Finish(err != nil)
				if err != nil {
					panic(err)
				}
			}()

			handle(sw, r, p)
		})
}

// Handler wraps the given Handler with stats reporting
func (m *HTTPMuxer) Handler(method, path string, handler http.Handler) {
	m.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
			handler.ServeHTTP(w, req)
		})
}

// HandlerFunc wraps the given HandlerFunc with stats reporting
func (m *HTTPMuxer) HandlerFunc(method, path string, handler http.HandlerFunc) {
	m.Handler(method, path, handler)
}

// Start timing the request
func (he *HTTPEndpoint) Start(w http.ResponseWriter) *HTTPResp {
	hr := httpRespPool.Get().(*HTTPResp)
	*hr = HTTPResp{
		ResponseWriter: w,
		statuses:       he.statuses,
		start:          time.Now(),
	}

	return hr
}

// Free releases this object and puts it back into the pool. This is optional
// and may only be used when you're sure no one is hanging onto the *HTTPResp.
func (hr *HTTPResp) Free() {
	httpRespPool.Put(hr)
}

// Finish records stats about this request.
func (hr *HTTPResp) Finish(paniced bool) time.Duration {
	dur := time.Now().Sub(hr.start)
	hr.statuses.record(hr.status, dur, paniced)
	return dur
}

// Status returns the http status that was sent, or 0 if none was sent
func (hr *HTTPResp) Status() int {
	return hr.status
}

// WriteHeader implements http.ResponseWriter.WriteHeader
func (hr *HTTPResp) WriteHeader(status int) {
	hr.status = status
	hr.ResponseWriter.WriteHeader(status)
}

// Hijack implements http.Hijacker
func (hr *HTTPResp) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := hr.ResponseWriter.(http.Hijacker)

	if !ok {
		err := fmt.Errorf("%T does not implement http.Hijacker", hr.ResponseWriter)
		return nil, nil, err
	}

	return hj.Hijack()
}

func (m *HTTPMuxer) newHTTPStatuses(name Name) *httpStatuses {
	newTimer := func(s string) *Timer {
		return NewTimer(name.Join(s), m.s.cfg.HTTPSamplePercent)
	}

	hss := &httpStatuses{
		all:    newTimer("all"),
		panics: newTimer("panic"),
		m:      map[int]*Timer{},
	}

	for _, code := range httpCodes {
		t := newTimer(fmt.Sprintf("%d", code))
		hss.m[code] = t
		hss.s = append(hss.s, t)
	}

	return hss
}

func (hss *httpStatuses) record(status int, dur time.Duration, paniced bool) {
	if status == 0 {
		status = http.StatusOK
	}

	var t *Timer
	if paniced {
		t = hss.panics
	} else {
		t = hss.m[status]
		if t == nil {
			t = hss.m[0]
		}
	}

	hss.all.Add(dur)
	t.Add(dur)
}

func (hss *httpStatuses) Snapshot(a Adder) {
	hss.all.snapshot(a, true)
	hss.panics.snapshot(a, true)
	for _, t := range hss.s {
		t.snapshot(a, true)
	}
}
