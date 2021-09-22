package service

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"

	git "github.com/go-git/go-git/v5"
	gitplumbing "github.com/go-git/go-git/v5/plumbing"
	gitfilemode "github.com/go-git/go-git/v5/plumbing/filemode"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	BasePath string
	repo     *Repo
}

func New(basePath string, repoURL string) (*Service, error) {
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
		repo:     r,
	}
	return s, nil
}

func (s *Service) GetFile(ctx context.Context, req *fspb.GetFileRequest) (*fspb.GetFileResponse, error) {
	req.Path = strings.TrimPrefix(req.Path, "/")

	commit, err := s.repo.repo.CommitObject(gitplumbing.NewHash(req.Commit))
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

	commit, err := s.repo.repo.CommitObject(gitplumbing.NewHash(req.Commit))
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
		res.Mode = fromGitFileMode(entry.Mode)
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
	res := &fspb.ListCommitsResponse{}
	iter, err := s.repo.repo.Log(&git.LogOptions{})
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

	res := &fspb.ListDirResponse{}

	commit, err := s.repo.repo.CommitObject(gitplumbing.NewHash(req.Commit))
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
			Mode: fromGitFileMode(treeEntry.Mode),
		})
	}
	if err != nil && err != io.EOF {
		return nil, status.Errorf(codes.Internal, "error while iterating through files for commit %q: %v", req.Commit, err)
	}

	return res, nil
}

func (s *Service) PushHook(w http.ResponseWriter, r *http.Request) {
	// Simply log the payload
	defer r.Body.Close()
	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("PushHook: Error reading body: %v", err)
		return
	}
	glog.Infof("PushHook: %s", string(contents))
}

func fromGitFileMode(m gitfilemode.FileMode) fspb.FileMode {
	switch m {
	case gitfilemode.Empty:
		return fspb.FileMode_MODE_EMPTY
	case gitfilemode.Dir:
		return fspb.FileMode_MODE_DIR
	case gitfilemode.Regular:
		return fspb.FileMode_MODE_REGULAR
	case gitfilemode.Deprecated:
		return fspb.FileMode_MODE_REGULAR
	case gitfilemode.Executable:
		return fspb.FileMode_MODE_EXECUTABLE
	case gitfilemode.Symlink:
		return fspb.FileMode_MODE_SYMLINK
	case gitfilemode.Submodule:
		return fspb.FileMode_MODE_SUBMODULE
	default:
		return fspb.FileMode_MODE_UNKNOWN
	}
}
