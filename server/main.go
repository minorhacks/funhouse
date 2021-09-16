package main

import (
	"flag"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/minorhacks/funhouse/service"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

var (
	httpPort = flag.Int("http_port", 8080, "Port of HTTP service")
	basePath = flag.String("base_path", "/tmp/funhouse", "Path to store cloned repository data")
)

func main() {
	flag.Parse()

	s := &service.Service{BasePath: *basePath}
	r := mux.NewRouter()
	r.HandleFunc("/hook/print", s.PrintHandler).Methods("POST")
	r.HandleFunc("/hook/mirror", s.MirrorHandler).Methods("POST")
	r.HandleFunc("/cat", s.CatHandler).Methods("GET")

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
