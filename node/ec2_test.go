package node

import (
	"net/http"
	"net/http/httptest"
)

func testEC2MetadataServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/latest/meta-data/local-ipv4",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("EC2PRIV"))
		})

	return httptest.NewServer(mux)
}
