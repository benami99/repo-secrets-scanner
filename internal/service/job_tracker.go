package service

import "sync"

var (
	jobMu    sync.Mutex
	jobState = map[string]string{} // key: "owner/repo" -> "running"|"done"
)

func SetJobState(owner, repo, state string) {
	jobMu.Lock()
	defer jobMu.Unlock()
	jobState[owner+"/"+repo] = state
}

func GetJobState(owner, repo string) string {
	jobMu.Lock()
	defer jobMu.Unlock()
	s, ok := jobState[owner+"/"+repo]
	if !ok {
		return "done" // default
	}
	return s
}
