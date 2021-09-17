package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/minorhacks/funhouse/service"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

var (
	httpPort   = flag.Int("http_port", 8080, "Port of HTTP service")
	basePath   = flag.String("base_path", "/tmp/funhouse", "Path to store cloned repository data")
	singleRepo = flag.String("single_repo", "", "If set, clone and serve a single repository at this URL")
)

func main() {
	flag.Parse()
	if err := app(); err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}

func app() error {
	var s *service.Service
	var err error
	if *singleRepo == "" {
		s, err = service.NewMulti(*basePath)
	} else {
		s, err = service.NewSingle(*basePath, *singleRepo)
	}
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/hook/print", s.PrintHandler).Methods("POST")
	r.HandleFunc("/cat", s.CatHandler).Methods("GET")
	r.HandleFunc("/getattr", s.GetAttrHandler).Methods("GET")
	if *singleRepo == "" {
		r.HandleFunc("/hook/mirror", s.MirrorHandler).Methods("POST")
	}

	addr := net.JoinHostPort("", strconv.FormatInt(int64(*httpPort), 10))
	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	glog.Infof("Listening on %s", addr)
	return srv.ListenAndServe()
}
