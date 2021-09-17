package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/minorhacks/funhouse/github"

	gitplumbing "github.com/go-git/go-git/v5/plumbing"
	"github.com/golang/glog"
	"github.com/kylelemons/godebug/pretty"
)

type Service struct {
	BasePath string
	repoMap  sync.Map
	singleRepo *Repo
}

func NewSingle(basePath string, repoURL string) (*Service, error) {
	r := &Repo{
		root: basePath,
		path: "",
	}
	err := r.init(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init repo: %v", err)
	}
	s := &Service{
		BasePath: basePath,
		singleRepo: r,
	}
	return s, nil
}

func NewMulti(basePath string) (*Service, error) {
	return &Service{
		BasePath: basePath,
	}, nil
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
	relPath := storePath(payload.Repository.URL)
	// Get directory for repository
	val, present := s.repoMap.LoadOrStore(relPath, &Repo{root: s.BasePath, path: relPath})
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

func (s *Service) CatHandler(w http.ResponseWriter, r *http.Request) {
	req := CatFileRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		glog.Errorf("%s: failed to unmarshal payload: %v", r.URL, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	val, ok := s.repoMap.Load(req.Repo)
	if !ok {
		glog.Errorf("%s: repo %q not found", r.URL, req.Repo)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	repo := val.(*Repo)
	commit, err := repo.repo.CommitObject(gitplumbing.NewHash(req.CommitHash))
	if err != nil {
		glog.Errorf("%s: can't get commit %q in repo %q: %v", r.URL, req.CommitHash, req.Repo, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	f, err := commit.File(req.Path)
	if err != nil {
		glog.Errorf("%s: can't get file %q in repo %s@%s: %v", r.URL, req.Path, req.Repo, req.CommitHash, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	rdr, err := f.Blob.Reader()
	if err != nil {
		glog.Errorf("%s: can't get reader for file %q in repo %s@%s: %v", r.URL, req.Path, req.Repo, req.CommitHash, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rdr.Close()

	_, err = io.Copy(w, rdr)
	if err != nil {
		glog.Errorf("%s: error copying %q in repo %s@%s: %v", r.URL, req.Path, req.Repo, req.CommitHash, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func storePath(urlStr string) string {
	parsed, _ := url.Parse(urlStr) // TODO: catch error
	return filepath.Join(parsed.Host, parsed.Path)
}
