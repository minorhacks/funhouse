package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/minorhacks/funhouse/github"

	git "github.com/go-git/go-git/v5"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/kylelemons/godebug/pretty"
)

var (
	httpPort = flag.Int("http_port", 8080, "Port of HTTP service")
	basePath = flag.String("base_path", "/tmp/funhouse", "Path to store cloned repository data")
)

func main() {
	flag.Parse()

	s := &Service{}
	r := mux.NewRouter()
	r.HandleFunc("/hook/print", s.PrintHandler).Methods("POST")
	r.HandleFunc("/hook/mirror", s.MirrorHandler).Methods("POST")

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

type Service struct {
	repoMap sync.Map
}

type Repo struct {
	mu   sync.RWMutex
	path string
  repo *git.Repository
}

func (s *Service) PrintHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Service) MirrorHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(1).Infof("%s at path: %s", r.Method, r.URL)

	// Decode webhook payload
	payload := github.PushPayload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		glog.Errorf("%s: failed to unmarshal payload: %v", r.URL, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get directory for repository
	val, present := s.repoMap.LoadOrStore(payload.Repository.URL, &Repo{path: filepath.Join(*basePath, storePath(payload.Repository.URL))})
	repo := val.(*Repo)
	// If not present, create and initialize it
	if !present {
		repo.mu.Lock()
		defer repo.mu.Unlock()

		var gitRepo *git.Repository
		if _, statErr := os.Stat(repo.path); os.IsNotExist(statErr) {
			gitRepo, err = git.PlainCloneContext(context.TODO(), repo.path /* isBare */, true, &git.CloneOptions{
				URL: payload.Repository.URL,
			})
			if err != nil {
				glog.Errorf("Failed to clone: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			glog.Infof("Successfully cloned %q to %q", payload.Repository.URL, repo.path)
		} else {
			gitRepo, err = git.PlainOpen(repo.path)
			if err != nil {
				glog.Error("Failed to open git repo: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			glog.Infof("Successfully opened repo at %q", repo.path)
		}

    repo.repo = gitRepo
	}

	// With directory locked
	// Pull new refs
}

func storePath(urlStr string) string {
	parsed, _ := url.Parse(urlStr) // TODO: catch error
	return fmt.Sprintf("%s/%s", parsed.Host, parsed.Path)
}
