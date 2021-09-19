package service

import (
	"context"
	"encoding/json"
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
	"github.com/kylelemons/godebug/pretty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	BasePath   string
	repoMap    sync.Map
	singleRepo *Repo
}

func (s *Service) GetFile(ctx context.Context, req *fspb.GetFileRequest) (*fspb.GetFileResponse, error) {
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
	req.Path = strings.TrimLeft(req.Path, "/")

	repo, err := s.getRepo("")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "repository not found")
	}

	commit, err := repo.repo.CommitObject(gitplumbing.NewHash(req.Commit))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "commit %q not found in repo: %v", req.Commit, err)
	}

	rootTree, err := commit.Tree()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get tree for commit %q: %v", req.Commit, err)
	}

	res := &fspb.GetAttributesResponse{
		AuthorTime: timestamppb.New(commit.Author.When),
		CommitTime: timestamppb.New(commit.Committer.When),
	}
	if req.Path == "" {
		// go-git doesn't have any concept of a "root directory" apparently, so
		// this is a hack to simulate it.
		res.Mode = fspb.FileMode_MODE_DIR
	} else {
		entry, err := rootTree.FindEntry(req.Path)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "can't get file %q at commit %q: %v", req.Path, req.Commit, err)
		}
		res.Mode = FromGitFileMode(entry.Mode)
		if entry.Mode.IsFile() {
			entryFile, err := rootTree.TreeEntryFile(entry)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "can't get TreeEntry for file %q at commit %q: %v", req.Path, req.Commit, err)
			}
			res.SizeBytes = uint64(entryFile.Blob.Size)
		}
	}

	return res, nil
}

func (s *Service) ListCommits(ctx context.Context, req *fspb.ListCommitsRequest) (*fspb.ListCommitsResponse, error) {
	repo, err := s.getRepo("")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "repository not found")
	}

	res := &fspb.ListCommitsResponse{}
	iter, err := repo.repo.Log(&git.LogOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get commit iterator: %v", err)
	}

	err = iter.ForEach(func(c *gitobject.Commit) error {
		res.Commits = append(res.Commits, c.Hash.String())
		return nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while traversing commits: %v", err)
	}

	return res, nil
}

func (s *Service) ListDir(ctx context.Context, req *fspb.ListDirRequest) (*fspb.ListDirResponse, error) {
	// Ensure the string is properly suffixed for a directory
	if !strings.HasPrefix(req.Path, "/") {
		req.Path = "/" + req.Path
	}
	if !strings.HasSuffix(req.Path, "/") {
		req.Path = req.Path + "/"
	}

	repo, err := s.getRepo("")
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "repository not found")
	}

	res := &fspb.ListDirResponse{}

	commit, err := repo.repo.CommitObject(gitplumbing.NewHash(req.Commit))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "commit %q not found in repo: %v", req.Commit, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get tree for commit %q: %v", req.Commit, err)
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
		res.Entries = append(res.Entries, &fspb.DirEntry{
			Name: delta,
			Mode: FromGitFileMode(treeEntry.Mode),
		})
	}
	if err != nil && err != io.EOF {
		return nil, status.Errorf(codes.Internal, "error while iterating through files for commit %q: %v", req.Commit, err)
	}

	return res, nil
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
