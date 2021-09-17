package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minorhacks/funhouse/github"

	gitplumbing "github.com/go-git/go-git/v5/plumbing"
	gitfilemode "github.com/go-git/go-git/v5/plumbing/filemode"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/golang/glog"
	"github.com/kylelemons/godebug/pretty"
)

type Service struct {
	BasePath   string
	repoMap    sync.Map
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
		BasePath:   basePath,
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

	repo, err := s.getRepo(req.Repo)
	if err != nil {
		glog.Errorf("%s: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
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

func (s *Service) GetAttrHandler(w http.ResponseWriter, r *http.Request) {
	req := GetAttrRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		glog.Errorf("%s: failed to unmarshal payload: %v", r.URL, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	path := strings.TrimLeft(req.Path, "/")

	repo, err := s.getRepo(req.Repo)
	if err != nil {
		glog.Errorf("%s: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	commit, err := repo.repo.CommitObject(gitplumbing.NewHash(req.CommitHash))
	if err != nil {
		glog.Errorf("%s: can't get commit %q in repo %q: %v", r.URL, req.CommitHash, req.Repo, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var f *gitobject.File
	if path == "" {
		// go-git doesn't have any concept of a "root directory" apparently, so
		// this is a hack to simulate it.
		f = &gitobject.File{
			Mode: gitfilemode.Dir,
		}
	} else {
		f, err = commit.File(path)
		if err != nil {
			glog.Errorf("%s: can't get file %q in repo %s@%s: %v", r.URL, path, req.Repo, req.CommitHash, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	res := &GetAttrResponse{}
	res.AuthorTime = JSONTime{commit.Author.When}
	res.CommitTime = JSONTime{commit.Committer.When}
	res.Size = uint64(f.Blob.Size)
	switch f.Mode {
	case gitfilemode.Empty:
		res.FileMode = FileModeEmpty
	case gitfilemode.Dir:
		res.FileMode = FileModeDir
	case gitfilemode.Regular:
		res.FileMode = FileModeRegular
	case gitfilemode.Deprecated:
		res.FileMode = FileModeRegular
	case gitfilemode.Executable:
		res.FileMode = FileModeExecutable
	case gitfilemode.Symlink:
		res.FileMode = FileModeSymlink
	case gitfilemode.Submodule:
		res.FileMode = FileModeSubmodule
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) getRepo(repo string) (*Repo, error) {
	if s.singleRepo != nil && repo != "" {
		return nil, fmt.Errorf("repo name must be unspecified for single-repo mode")
	}
	if s.singleRepo == nil && repo == "" {
		return nil, fmt.Errorf("repo name must be specified for multi-repo mode")
	}
	if s.singleRepo != nil {
		return s.singleRepo, nil
	}
	val, ok := s.repoMap.Load(repo)
	if !ok {
		return nil, fmt.Errorf("repo %q not found", repo)
	}
	return val.(*Repo), nil
}

func storePath(urlStr string) string {
	parsed, _ := url.Parse(urlStr) // TODO: catch error
	return filepath.Join(parsed.Host, parsed.Path)
}
