package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/golang/glog"
)

type Repo struct {
	mu   sync.RWMutex
	root string
	path string
	repo *git.Repository
}

func (r *Repo) fullPath() string {
	return filepath.Join(r.root, r.path)
}

func (r *Repo) init(url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.repo == nil {
		var gitRepo *git.Repository
		var err error
		if _, statErr := os.Stat(r.path); os.IsNotExist(statErr) {
			gitRepo, err = git.PlainCloneContext(context.TODO(), r.fullPath(),/* isBare */ true, &git.CloneOptions{
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
		RefSpecs: []gitconfig.RefSpec{gitconfig.RefSpec(fmt.Sprintf("+%s:%s", ref, ref))},
		Force:    true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull ref %q for repo %q: %v", ref, r.path, err)
	}
	return nil
}
