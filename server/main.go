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
	"github.com/go-git/go-git/v5/config"
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

func (r *Repo) init(url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.repo == nil {
		var gitRepo *git.Repository
		var err error
		if _, statErr := os.Stat(r.path); os.IsNotExist(statErr) {
			gitRepo, err = git.PlainCloneContext(context.TODO(), r.path /* isBare */, true, &git.CloneOptions{
				URL: url,
			})
			if err != nil {
				return fmt.Errorf("failed to clone: %v", err)
			}
			glog.Infof("Successfully cloned %q to %q", url, r.path)
		} else {
			gitRepo, err = git.PlainOpen(r.path)
			if err != nil {
				return fmt.Errorf("failed to open existing git repo: %v", err)
			}
			glog.Infof("Successfully opened repo at %q", r.path)
		}
		r.repo = gitRepo
	}
	return nil
}

func (r *Repo) pull(ref string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("+%s:%s", ref, ref))},
		Force:    true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull ref %q for repo %q: %v", ref, r.path, err)
	}
	return nil
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
		err := repo.init(payload.Repository.URL)
		if err != nil {
			glog.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = repo.pull(payload.Ref)
	if err != nil {
		glog.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func storePath(urlStr string) string {
	parsed, _ := url.Parse(urlStr) // TODO: catch error
	return fmt.Sprintf("%s/%s", parsed.Host, parsed.Path)
}
