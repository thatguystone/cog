package stats

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/thatguystone/cog/clog"
)

// HTTPMuxer is a wrapper around httprouter.Router. It intercepts all requests
// and reports useful information about them.
type HTTPMuxer struct {
	// Exposed so that you can set Options on it, use it as the server
	// handler, etc. Any routes added directly here will have no stats
	// recorded for them.
	R    *httprouter.Router
	log  *clog.Logger
	name string
	s    *S
}

type httpStatuses struct {
	all    *Timer
	panics *Timer
	m      map[int]*Timer
	s      []*Timer
}

type httpResp struct {
	http.ResponseWriter
	status int
}

var httpCodes = []int{
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

// NewHTTPMuxer creates a new HTTP muxer that exposes status information for
// all endpoints. Each endpoint has a corresponding /_status/<URL>. At
// /_status, all endpoints are reported.
func (s *S) NewHTTPMuxer(name string) *HTTPMuxer {
	m := &HTTPMuxer{
		name: name,
		log:  s.log.Get("http"),
		s:    s,
		R:    httprouter.New(),
	}

	return m
}

// Handle wraps the given Handle with stats reporting
func (m *HTTPMuxer) Handle(method, path string, handle httprouter.Handle) {
	path = httprouter.CleanPath(path)
	name := Join(m.name, path, method)

	status := m.newHTTPStatuses(name)
	m.s.AddSnapshotter(name, status)

	m.R.Handle(method, path,
		func(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
			start := time.Now()

			hrw := &httpResp{ResponseWriter: rw}
			defer func() {
				dur := time.Now().Sub(start)

				err := recover()
				status.record(hrw.status, dur, err != nil)

				if err != nil {
					panic(err)
				}
			}()

			rw = hrw
			handle(rw, req, p)
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

func (rw *httpResp) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (m *HTTPMuxer) newHTTPStatuses(name string) *httpStatuses {
	newTimer := func(s string) *Timer {
		return NewTimer(Join(name, s), m.s.cfg.HTTPSamplePercent)
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
	hss.panics.snapshot(a, true)
	for _, t := range hss.s {
		t.snapshot(a, true)
	}
}
