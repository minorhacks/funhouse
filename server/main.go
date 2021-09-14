package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"strconv"
	"time"

  "github.com/minorhacks/funhouse/github"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/kylelemons/godebug/pretty"
)

var (
	httpPort = flag.Int("http_port", 8080, "Port of HTTP service")
)

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/hook/print", PrintHandler).Methods("POST")
  r.HandleFunc("/hook/mirror", MirrorHandler).Methods("POST")

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

	defer r.Body.Close()
	payload := github.PushPayload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		glog.Errorf("%s: failed to unmarshal payload: %v", r.URL, err)
		w.WriteHeader(http.StatusBadRequest)
    return
	}

	pretty.Print(payload)
}

func MirrorHandler(w http.ResponseWriter, r *http.Request) {
  glog.V(1).Infof("%s at path: %s", r.Method, r.URL)
}
