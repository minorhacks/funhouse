package main

import (
	"flag"
	"net"
	"net/http"
	"strconv"
	"time"

  "github.com/golang/glog"
	"github.com/gorilla/mux"
)

var (
	httpPort = flag.Int("http_port", 8080, "Port of HTTP service")
)

func main() {
  flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/hook/print", PrintHandler).Methods("POST")

  addr := net.JoinHostPort("", strconv.FormatInt(int64(*httpPort), 10))
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
  glog.Infof("Listening on %s", addr)
	glog.Fatal(srv.ListenAndServe())
}

func PrintHandler(w http.ResponseWriter, r *http.Request) {
  glog.V(1).Infof("%s at path: %s", r.Method, r.URL)
}
