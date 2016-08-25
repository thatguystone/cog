package eio

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iheartradio/cog/check"
)

type httpTest struct {
	c    *check.C
	srv  *httptest.Server
	p    Producer
	reqs chan httpReq
}

type httpReq struct {
	req  *http.Request
	body []byte
}

func newHTTPTest(t *testing.T, args Args) (*check.C, *httpTest) {
	mux := http.NewServeMux()
	ht := &httpTest{
		c:    check.New(t),
		srv:  httptest.NewServer(mux),
		reqs: make(chan httpReq, 8),
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		ht.c.MustNotError(err)
		ht.reqs <- httpReq{
			req:  r,
			body: body,
		}
	}

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/some/path", handler)

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	var err error
	args["Servers"] = []string{ht.srv.URL}
	ht.p, err = NewProducer("http", args)
	ht.c.MustNotError(err)

	return ht.c, ht
}

func (ht *httpTest) exit() {
	ht.srv.Close()
	ht.p.Close()
}

func TestHTTPBasic(t *testing.T) {
	ep := "some/path"
	c, ht := newHTTPTest(t, Args{
		"Endpoint":          ep,
		"BatchDelay":        "100us",
		"InitialRetryDelay": "100us",
		"MaxRetryBackoff":   "4ms",
	})
	defer ht.exit()

	c.NotError(ht.p.Rotate())

	for i := 0; i < 10; i++ {
		ht.p.Produce([]byte("test"))
	}

	lines := 0
	drain := func() {
		for {
			select {
			case r := <-ht.reqs:
				c.Equal(r.req.URL.Path, "/"+ep)
				c.Equal(r.req.Header.Get("Content-Type"), httpContentType)
				c.Equal(r.req.Header.Get("Content-Length"), "")
				lines += bytes.Count(r.body, []byte("\n"))

			default:
				return
			}
		}
	}

	c.Until(time.Second, func() bool {
		drain()
		return lines == 10
	})
}

func TestHTTPDisableChunked(t *testing.T) {
	c, ht := newHTTPTest(t, Args{
		"BatchDelay":     "100us",
		"MimeType":       "application/json",
		"DisableChunked": true,
	})
	defer ht.exit()

	ht.p.Produce([]byte("test"))

	select {
	case r := <-ht.reqs:
		c.NotEqual(r.req.Header.Get("Content-Length"), "")

	case <-time.After(time.Second):
		c.Fatal("did not get request after 1 second")
	}
}

func TestHTTPContentType(t *testing.T) {
	c, ht := newHTTPTest(t, Args{
		"BatchDelay": "100us",
		"MimeType":   "application/json",
	})
	defer ht.exit()

	ht.p.Produce([]byte("test"))

	select {
	case r := <-ht.reqs:
		c.Equal(r.req.Header.Get("Content-Type"), "application/json")

	case <-time.After(time.Second):
		c.Fatal("did not get request after 1 second")
	}
}

func TestHTTPScheming(t *testing.T) {
	c := check.New(t)

	p, err := regdPs["http"](Args{
		"Servers": []string{
			"12345",
			"http://12345",
			"https://blah",
		},
	})
	c.MustNotError(err)

	for _, s := range p.(*HTTPProducer).Args.Servers {
		c.Equal(
			strings.Count(s, "http://")+
				strings.Count(s, "https://"),
			1, "invalid: %s", s)
		c.True(
			strings.HasPrefix(s, "http://") ||
				strings.HasPrefix(s, "https://"), "not set: %s", s)
	}
}

func TestHTTPCloseThenSend(t *testing.T) {
	_, ht := newHTTPTest(t, Args{})
	defer ht.exit()

	// Don't panic!
	ht.p.Produce([]byte("test"))
}

func TestHTTPSizedFlush(t *testing.T) {
	c, ht := newHTTPTest(t, Args{
		"BatchSize":  4,
		"BatchDelay": 0,
	})
	defer ht.exit()

	for i := 0; i < 3; i++ {
		ht.p.Produce([]byte("test"))
	}

	select {
	case <-ht.reqs:
		c.Fatal("should not have flushed!")
	case <-time.After(time.Millisecond * 5):
	}

	for i := 0; i < 3; i++ {
		ht.p.Produce([]byte("test"))
	}

	select {
	case r := <-ht.reqs:
		c.Equal(4, bytes.Count(r.body, []byte("\n")))
	case <-time.After(time.Second):
		c.Fatal("no request after 1s")
	}
}

func TestHTTPRetryTimeout(t *testing.T) {
	c, ht := newHTTPTest(t, Args{
		"Endpoint":          "error",
		"BatchDelay":        "100us",
		"InitialRetryDelay": "100us",
		"MaxRetryBackoff":   "1ms",
	})
	defer ht.exit()

	ht.p.Produce([]byte("test"))

	select {
	case err := <-ht.p.Errs():
		c.Error(err)
	case <-time.After(time.Second * 1):
		c.Fatal("did not get error after 1s")
	}
}

func TestHTTPError(t *testing.T) {
	c := check.New(t)

	_, err := NewProducer("http", Args{})
	c.Error(err)
}
