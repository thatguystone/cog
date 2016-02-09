package stats

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/thatguystone/cog/check/chlog"
)

func TestHTTPBasic(t *testing.T) {
	c, clog := chlog.New(t)
	s := NewS(Config{}, clog.Get("stats"))
	mux := s.NewHTTPMuxer("http")

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

	srv := httptest.NewServer(mux.R)
	defer srv.Close()

	wg := sync.WaitGroup{}
	get := func(url string) {
		defer wg.Done()

		resp, err := http.Get(url)
		c.MustNotError(err)
		resp.Body.Close()
	}

	for i := 0; i < 10; i++ {
		wg.Add(5)
		go get(fmt.Sprintf("%s/sleep/1/", srv.URL))
		go get(fmt.Sprintf("%s/sleep/5/", srv.URL))
		go get(fmt.Sprintf("%s/404", srv.URL))
		go get(fmt.Sprintf("%s/500", srv.URL))
		go get(fmt.Sprintf("%s/164", srv.URL))
	}

	wg.Wait()

	snap := s.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name, st.Val)
	}

	c.Equal(snap.Get("http./sleep/1.GET.200.count").Val.(int64), 10)
	c.Equal(snap.Get("http./500.GET.500.count").Val.(int64), 10)
	c.Equal(snap.Get("http./404.GET.404.count").Val.(int64), 10)
	c.Equal(snap.Get("http./164.GET.0.count").Val.(int64), 10)
}

func TestHTTPPanic(t *testing.T) {
	c, clog := chlog.New(t)
	s := NewS(Config{}, clog.Get("stats"))
	mux := s.NewHTTPMuxer("http")

	mux.HandlerFunc("GET", "/panic",
		func(rw http.ResponseWriter, req *http.Request) {
			panic("i give up")
		})

	h, p, _ := mux.R.Lookup("GET", "/panic")
	c.MustTrue(h != nil)

	c.Panic(func() {
		rw := httptest.NewRecorder()
		h(rw, nil, p)
	})

	snap := s.snapshot()
	for _, st := range snap {
		c.Logf("%s = %v", st.Name, st.Val)
	}

	c.Equal(snap.Get("http./panic.GET.panic.count").Val.(int64), 1)
}
