package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/benami99/repo-scanner/internal/model"
	"github.com/benami99/repo-scanner/internal/repository"
)

type RepositoryClient interface {
	ListBranches(owner, repo string) ([]string, error)
	ListCommits(owner, repo, branch string) ([]model.Commit, error)
	GetCommitDiff(owner, repo, sha string) (map[string]string, error)
}

type Scanner struct {
	store  repository.Repository
	client RepositoryClient
}

func NewScanner(store repository.Repository, client RepositoryClient) *Scanner {
	return &Scanner{store: store, client: client}
}

func (s *Scanner) ScanRepo(ctx context.Context, owner, repo, fromSHA string) error {
	r := model.Repo{Owner: owner, Name: repo}

	// ✅ Resume support
	if fromSHA == "" {
		fromSHA = s.store.GetLastScannedSHA(r)
	}

	// Get all branches
	branches, err := s.client.ListBranches(owner, repo)
	if err != nil {
		return fmt.Errorf("list branches: %w", err)
	}

	//Get all commits across all branches
	var allCommits []model.Commit
	for _, b := range branches {
		cs, err := s.client.ListCommits(owner, repo, b)
		if err != nil {
			fmt.Printf("warn: failed to list commits for branch %s: %v\n", b, err)
			continue
		}
		allCommits = append(allCommits, cs...)
	}

	// Dedup commits
	commitMap := make(map[string]model.Commit)
	for _, c := range allCommits {
		commitMap[c.SHA] = c
	}

	// Create a slice of all unique commits
	uniqCommits := make([]model.Commit, 0, len(commitMap))
	for _, c := range commitMap {
		uniqCommits = append(uniqCommits, c)
	}

	// Sort all commits descending by commit time
	sort.Slice(uniqCommits, func(i, j int) bool {
		return uniqCommits[i].Timestamp.After(uniqCommits[j].Timestamp)
	})

	// Iterate all sorted commits.
	// For each commit find its file names and diffs by SHA
	// Iterate all diffs and find leaks
	started := fromSHA == "" // start from the first commit in the list of fromSHA none specified
	for _, c := range uniqCommits {
		// check cancel from client
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// ✅ Resume from last SHA
		if !started {
			if c.SHA == fromSHA {
				started = true
			}
			continue
		}

		// for each commit - get all files
		files, err := s.client.GetCommitDiff(owner, repo, c.SHA)
		if err != nil {
			fmt.Printf("warn: failed diff for %s: %v\n", c.SHA, err)
			continue
		}

		// ✅ Scan all file diffs. "files" is a map fileName --> Diff
		for path, patch := range files {
			leaks := detectLeaks(patch)
			for _, l := range leaks {
				finding := model.Finding{
					Committer: c.Committer,
					CommitSHA: c.SHA,
					FilePath:  path,
					LeakType:  l.Type,
					LeakValue: l.Value,
					Timestamp: c.Timestamp,
				}
				// Save the finding in memory
				if err := s.store.SaveFinding(r, finding); err != nil {
					fmt.Printf("warn: save finding failed: %v\n", err)
				} else {
					fmt.Printf("✅ found: %s %s %s\n", c.SHA[:8], path, l.Value)
				}
			}
		}

		// Persist progress after each commit, even if no findings were found
		s.store.SaveLastScannedSHA(r, c.SHA)

		time.Sleep(200 * time.Millisecond) // avoid hitting rate limits
	}

	return nil
}

func (s *Scanner) ListFindings(owner, repo string) []model.Finding {
	return s.store.ListFindings(model.Repo{
		Owner: owner,
		Name:  repo,
	})
}
