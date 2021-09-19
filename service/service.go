package service

import (
	"encoding/json"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minorhacks/funhouse/github"
	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"

	git "github.com/go-git/go-git/v5"
	gitplumbing "github.com/go-git/go-git/v5/plumbing"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/golang/glog"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"github.com/kylelemons/godebug/pretty"
)

type Service struct {
	BasePath   string
	repoMap    sync.Map
	singleRepo *Repo
}

func (s *Service) GetFile(ctx context.Context, req *fspb.GetFileRequest) (*fspb.GetFileResponse, error ){
	req.Path = strings.TrimPrefix(req.Path, "/")

	repo, err := s.getRepo("")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "repository not found")
	}
	commit, err := repo.repo.CommitObject(gitplumbing.NewHash(req.Commit))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "commit %q not found in repo: %v", req.Commit, err)
	}
	f, err := commit.File(req.Path)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "file %q not found at commit %q: %v", req.Path, req.Commit, err)
	}
	rdr, err := f.Blob.Reader()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get reader for file %q at commit %q: %v", req.Path, req.Commit, err)
	}
	defer rdr.Close()

	res := &fspb.GetFileResponse{}
	res.Contents, err = ioutil.ReadAll(rdr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error copying from %q at commit %q: %v", req.Path, req.Commit, err)
	}
	return res, nil
}

func (s *Service) GetAttributes(ctx context.Context, req *fspb.GetAttributesRequest) (*fspb.GetAttributesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetAttributes() not yet implemented")
}

func (s *Service) ListCommits(ctx context.Context, req *fspb.ListCommitsRequest) (*fspb.ListCommitsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListCommits() not yet implemented")
}

func (s *Service) ListDir(ctx context.Context, req *fspb.ListDirRequest) (*fspb.ListDirResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListDir() not yet implemented")
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

	rootTree, err := commit.Tree()
	if err != nil {
		glog.Errorf("%s: can't get tree for commit %q: %v", r.URL, req.CommitHash, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := &GetAttrResponse{}
	res.AuthorTime = JSONTime{commit.Author.When}
	res.CommitTime = JSONTime{commit.Committer.When}
	if path == "" {
		// go-git doesn't have any concept of a "root directory" apparently, so
		// this is a hack to simulate it.
		res.FileMode = FileModeDir
	} else {
		entry, err := rootTree.FindEntry(path)
		if err != nil {
			glog.Errorf("%s: can't get file %q in repo %s@%s: %v", r.URL, path, req.Repo, req.CommitHash, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		res.FileMode = FromGitFileMode(entry.Mode)
		if entry.Mode.IsFile() {
			entryFile, err := rootTree.TreeEntryFile(entry)
			if err != nil {
				glog.Errorf("%s: can't get file %q in repo %s@%s: %v", r.URL, path, req.Repo, req.CommitHash, err)
				w.WriteHeader(http.StatusNotFound)
				return
			}
			res.Size = uint64(entryFile.Blob.Size)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) CommitsHandler(w http.ResponseWriter, r *http.Request) {
	req := ListCommitsRequest{}
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

	res := &ListCommitsResponse{}
	iter, err := repo.repo.Log(&git.LogOptions{})
	if err != nil {
		glog.Errorf("%s: failed to get commit iterator: %v", r.URL, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = iter.ForEach(func(c *gitobject.Commit) error {
		res.CommitHashes = append(res.CommitHashes, c.Hash.String())
		return nil
	})
	if err != nil {
		glog.Errorf("%s: error while traversing commits: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) ListDirHandler(w http.ResponseWriter, r *http.Request) {
	req := ListDirRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		glog.Errorf("%s: failed to unmarshal payload: %v", r.URL, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Ensure the string is properly suffixed for a directory
	if !strings.HasPrefix(req.Path, "/") {
		req.Path = "/" + req.Path
	}
	if !strings.HasSuffix(req.Path, "/") {
		req.Path = req.Path + "/"
	}

	repo, err := s.getRepo(req.Repo)
	if err != nil {
		glog.Errorf("%s: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	res := &ListDirResponse{}

	commit, err := repo.repo.CommitObject(gitplumbing.NewHash(req.CommitHash))
	if err != nil {
		glog.Errorf("%s: can't get commit %q in repo %q: %v", r.URL, req.CommitHash, req.Repo, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		glog.Errorf("%s: can't get tree for commit %q: %v", r.URL, req.CommitHash, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	treeIter := gitobject.NewTreeWalker(tree, true /* recursive */, nil)
	defer treeIter.Close()

	var name string
	var treeEntry gitobject.TreeEntry
	for name, treeEntry, err = treeIter.Next(); err == nil; name, treeEntry, err = treeIter.Next() {
		name = "/" + name
		// Filter out all files not direct descendants of the specified dir
		if !strings.HasPrefix(name, req.Path) {
			continue
		}
		delta := strings.TrimPrefix(name, req.Path)
		if strings.Contains(delta, "/") {
			continue
		}
		// Return an entry for the remaining entries
		res.Entries = append(res.Entries, &DirEntry{
			Name:     delta,
			FileMode: FromGitFileMode(treeEntry.Mode),
		})
	}
	if err != nil && err != io.EOF {
		glog.Errorf("%s: error while iterating through files for commit %q: %v", r.URL, req.CommitHash, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
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
