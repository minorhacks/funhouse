package service

import (
	"fmt"
	"strings"
	"time"
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
