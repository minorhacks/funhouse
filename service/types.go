package service

import (
	"fmt"
	"strings"
	"syscall"
	"time"

	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"

	gitfilemode "github.com/go-git/go-git/v5/plumbing/filemode"
)

type FileModeOld uint32

const (
	FileModeEmpty FileModeOld = iota
	FileModeDir
	FileModeRegular
	FileModeExecutable
	FileModeSymlink
	FileModeSubmodule
)

func FromGitFileMode(m gitfilemode.FileMode) fspb.FileMode {
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

type JSONTime struct {
	time.Time
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.Format(time.RFC3339Nano))), nil
}

func (t JSONTime) UnmarshalJSON(data []byte) error {
	var err error
	t.Time, err = time.Parse(time.RFC3339Nano, strings.Trim(string(data), `"`))
	return err
}

type CatFileRequest struct {
	Repo       string `json:"repo"`
	CommitHash string `json:"commit_hash"`
	Path       string `json:"path"`
}

type GetAttrRequest struct {
	Repo       string `json:"repo"`
	CommitHash string `json:"commit_hash"`
	Path       string `json:"path"`
}

type GetAttrResponse struct {
	FileMode   FileModeOld `json:"file_mode"`
	Size       uint64      `json:"size"`
	CommitTime JSONTime    `json:"commit_time"`
	AuthorTime JSONTime    `json:"author_time"`
}

type ListCommitsRequest struct {
	Repo string `json:"repo"`
}

type ListCommitsResponse struct {
	CommitHashes []string `json:"commit_hashes"`
}

type ListDirRequest struct {
	Repo       string `json:"repo"`
	CommitHash string `json:"commit_hash"`
	Path       string `json:"path"`
}

type ListDirResponse struct {
	Entries []*DirEntry `json:"entries"`
}

type DirEntry struct {
	Name     string      `json:"name"`
	FileMode FileModeOld `json:"file_mode"`
}
