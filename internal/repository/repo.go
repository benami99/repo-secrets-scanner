package repository

import (
	"fmt"
	"sync"

	"github.com/benami99/repo-scanner/internal/model"
)

// Repository is a very small abstraction for storing findings.
type Repository interface {
	SaveFinding(r model.Repo, f model.Finding) error
	ListFindings(r model.Repo) []model.Finding
	GetLastScannedSHA(r model.Repo) string
	SaveLastScannedSHA(r model.Repo, sha string)
}

type memoryRepo struct {
	mu          sync.Mutex
	findings    map[string][]model.Finding
	lastScanned map[string]string // repoKey -> last processed SHA
}

func NewMemoryStore() Repository {
	return &memoryRepo{findings: make(map[string][]model.Finding),
		lastScanned: make(map[string]string),
	}

}

// keyFor generates a unique key for a given model.Repo.
// It is used internally for storing findings in the memory repository.
//
// The key is a string in the format "owner/name".
//
// For example, if the given model.Repo has an owner of "foo" and a name of "bar",
// the resulting key would be "foo/bar".
//
// This function is not intended to be used externally.
func keyFor(r model.Repo) string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// SaveFinding Save the finding in a map of key (onwer+repo) --> finding entity
func (m *memoryRepo) SaveFinding(r model.Repo, f model.Finding) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := keyFor(r)
	// prevent duplicates of the same finding tuple
	for _, existing := range m.findings[k] {
		if existing.CommitSHA == f.CommitSHA && existing.FilePath == f.FilePath && existing.LeakType == f.LeakType && existing.LeakValue == f.LeakValue {
			return nil
		}
	}
	m.findings[k] = append(m.findings[k], f)
	// also update last scanned to allow resume even if SaveLastScannedSHA isn't called
	m.lastScanned[k] = f.CommitSHA
	return nil
}

func (m *memoryRepo) ListFindings(r model.Repo) []model.Finding {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.findings[keyFor(r)]
}

func (m *memoryRepo) GetLastScannedSHA(r model.Repo) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastScanned[keyFor(r)]
}

// SaveLastScannedSHA explicitly persists the last processed commit SHA for a repo.
func (m *memoryRepo) SaveLastScannedSHA(r model.Repo, sha string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastScanned[keyFor(r)] = sha
}
