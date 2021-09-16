package service

type CatFileRequest struct {
	Repo string `json:"repo"`
	Path string `json:"path"`
	CommitHash string `json:"commit_hash"`
}