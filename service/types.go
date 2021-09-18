package service

import (
	"fmt"
	"strings"
	"time"

	gitfilemode "github.com/go-git/go-git/v5/plumbing/filemode"
)

type FileMode uint32

const (
	FileModeEmpty FileMode = iota
	FileModeDir
	FileModeRegular
	FileModeExecutable
	FileModeSymlink
	FileModeSubmodule
)

func FromGitFileMode(m gitfilemode.FileMode) FileMode {
	switch m {
	case gitfilemode.Empty:
		return FileModeEmpty
	case gitfilemode.Dir:
		return FileModeDir
	case gitfilemode.Regular:
		return FileModeRegular
	case gitfilemode.Deprecated:
		return FileModeRegular
	case gitfilemode.Executable:
		return FileModeExecutable
	case gitfilemode.Symlink:
		return FileModeSymlink
	case gitfilemode.Submodule:
		return FileModeSubmodule
	default:
		panic(fmt.Sprintf("Unhandled filemode: %v", m))
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
	FileMode   FileMode `json:"file_mode"`
	Size       uint64   `json:"size"`
	CommitTime JSONTime `json:"commit_time"`
	AuthorTime JSONTime `json:"author_time"`
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
	Name       string   `json:"name"`
	FileMode   FileMode `json:"file_mode"`
	Size       uint64   `json:"size"`
	CommitTime JSONTime `json:"commit_time"`
	AuthorTime JSONTime `json:"author_time"`
}
