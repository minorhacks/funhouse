package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/minorhacks/funhouse/github"

	"github.com/golang/glog"
	"github.com/kylelemons/godebug/pretty"
)

type Service struct {
	BasePath string
	repoMap sync.Map
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
	val, present := s.repoMap.LoadOrStore(payload.Repository.URL, &Repo{path: filepath.Join(s.BasePath, storePath(payload.Repository.URL))})
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

}

func storePath(urlStr string) string {
	parsed, _ := url.Parse(urlStr) // TODO: catch error
	return fmt.Sprintf("%s/%s", parsed.Host, parsed.Path)
}