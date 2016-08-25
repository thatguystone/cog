package statc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/iheartradio/cog/check"
)

const statusKey = "test-key"

func newHTTPTest(t *testing.T) (*check.C, *sTest, *HTTPMuxer) {
	c, st := newTest(t, &Config{
		StatusKey:           statusKey,
		MemStatsInterval:    -1,
		disableRuntimeStats: true,
	})
	mux := st.NewHTTPMuxer("http")
	return c, st, mux
}

func TestHTTPBasic(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	mux.HandlerFunc("GET", "/sleep/1",
		func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond)
		})
	mux.HandlerFunc("GET", "/sleep/5",
		func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(time.Millisecond * 5)
		})
	mux.HandlerFunc("GET", "/404",
		func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "404", 404)
		})
	mux.HandlerFunc("GET", "/500",
		func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "500", 500)
		})
	mux.HandlerFunc("GET", "/164",
		func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "164", 164)
		})
	mux.HandlerFunc("GET", "/hijack",
		func(rw http.ResponseWriter, req *http.Request) {
			c, _, err := rw.(http.Hijacker).Hijack()
			if err != nil {
				rw.WriteHeader(500)
				rw.Write([]byte(err.Error()))
			} else {
				c.Close()
			}
		})

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	wg := sync.WaitGroup{}
	get := func(url string, expectOk bool) {
		defer wg.Done()

		resp, err := http.Get(url)

		if expectOk {
			c.MustNotError(err)
			resp.Body.Close()
		} else {
			c.MustError(err)
		}
	}

	for i := 0; i < 10; i++ {
		wg.Add(6)
		go get(fmt.Sprintf("%s/sleep/1/", srv.URL), true)
		go get(fmt.Sprintf("%s/sleep/5/", srv.URL), true)
		go get(fmt.Sprintf("%s/404", srv.URL), true)
		go get(fmt.Sprintf("%s/500", srv.URL), true)
		go get(fmt.Sprintf("%s/164", srv.URL), true)
		go get(fmt.Sprintf("%s/hijack", srv.URL), false)
	}

	wg.Wait()

	rec := httptest.NewRecorder()
	h, params, _ := mux.R.Lookup("GET", "/hijack")
	c.MustTrue(h != nil)
	h(rec, nil, params)
	c.Equal(rec.Code, 500)
	c.Contains(rec.Body.String(), "does not implement http.Hijacker")

	snap := st.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name, st.Val)
	}

	c.Equal(snap.Get(st.Names("http", "/sleep/1", "GET", "all", "count")).Val.(int64), 10)
	c.Equal(snap.Get(st.Names("http", "/sleep/1", "GET", "200", "count")).Val.(int64), 10)
	c.Equal(snap.Get(st.Names("http", "/500", "GET", "500", "count")).Val.(int64), 10)
	c.Equal(snap.Get(st.Names("http", "/404", "GET", "404", "count")).Val.(int64), 10)
	c.Equal(snap.Get(st.Names("http", "/164", "GET", "0", "count")).Val.(int64), 10)
}

func TestHTTPPanic(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	mux.HandlerFunc("GET", "/panic",
		func(rw http.ResponseWriter, req *http.Request) {
			panic("i give up")
		})

	h, p, _ := mux.R.Lookup("GET", "/panic")
	c.MustTrue(h != nil)

	c.Panics(func() {
		rw := httptest.NewRecorder()
		h(rw, nil, p)
	})

	snap := st.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name, st.Val)
	}

	c.Equal(snap.Get(st.Names("http", "/panic", "GET", "panic", "count")).Val.(int64), 1)
}

func TestHTTPStatusHandler(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	st.NewTimer("some.timer", 100).Add(time.Second)
	st.NewCounter("module.counter", false).Add(100)
	st.NewGauge("my.gauge").Set(9)
	st.NewStringGauge("str.gauge").Set("some string")
	st.doSnapshot()

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/_status?key=" + statusKey)
	c.MustNotError(err)
	defer resp.Body.Close()

	r, err := ioutil.ReadAll(resp.Body)
	c.MustNotError(err)

	out := `{
	"module": {
		"counter": 100
	},
	"my": {
		"gauge": 9
	},
	"some": {
		"timer": {
			"count": 1,
			"max": 1000000000,
			"mean": 1000000000,
			"min": 1000000000,
			"p50": 1000000000,
			"p75": 1000000000,
			"p90": 1000000000,
			"p95": 1000000000,
			"stddev": 0
		}
	},
	"str": {
		"gauge": "some string"
	}` + "\n}"

	c.Equal(string(r), out)
}

func TestHTTPStatusHandlerError(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	st.lastSnap = Snapshot{
		Stat{
			Name: "blah",
			Val:  nil,
		},
	}

	resp, err := http.Get(srv.URL + "/_status?key=" + statusKey)
	c.MustNotError(err)
	defer resp.Body.Close()
	c.Equal(resp.StatusCode, http.StatusInternalServerError)

	resp, err = http.Get(srv.URL + "/_status?key=invalidkey")
	c.MustNotError(err)
	defer resp.Body.Close()
	c.Equal(resp.StatusCode, http.StatusForbidden)
}

func TestHTTPCustomHandlers(t *testing.T) {
	c, st, mux := newHTTPTest(t)
	defer st.exit.Exit()

	ep := mux.Endpoint("GET", "/custom")
	mux.R.Handle("GET", "/custom",
		func(rw http.ResponseWriter, req *http.Request, _ httprouter.Params) {
			w := ep.Start(rw)
			w.WriteHeader(404)
			c.Equal(w.Status(), 404)

			w.Finish(false)
			w.Free()
		})

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/custom")
	c.MustNotError(err)
	defer resp.Body.Close()

	c.Equal(resp.StatusCode, 404)
}
