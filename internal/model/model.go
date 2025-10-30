package model

import "time"

type Repo struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type Commit struct {
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Committer string    `json:"committer"`
	Timestamp time.Time `json:"timestamp"`
}

type Finding struct {
	CommitSHA string    `json:"commit_sha"`
	Committer string    `json:"committer"`
	FilePath  string    `json:"file_path"`
	LeakType  string    `json:"leak_type"`
	LeakValue string    `json:"leak_value"`
	Timestamp time.Time `json:"timestamp"`
}
