package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/benami99/repo-secrets-scanner/internal/model"
)

type GithubClient struct {
	token  string
	client *http.Client
}

func NewGithubClient(token string) *GithubClient {
	return &GithubClient{token: token, client: &http.Client{Timeout: 15 * time.Second}}
}

func (g *GithubClient) ListBranches(owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches?per_page=100", owner, repo)

	var all []string

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		if g.token != "" {
			req.Header.Set("Authorization", "token "+g.token)
		}
		resp, err := g.doRequest(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("github branches: %d %s", resp.StatusCode, string(b))
		}
		var raw []struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return nil, err
		}

		for _, r := range raw {
			all = append(all, r.Name)
		}

		url = getNextPage(resp)
	}
	return all, nil
}

func (g *GithubClient) ListCommits(owner, repo, branch string) ([]model.Commit, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?sha=%s&per_page=100", owner, repo, branch)

	var commits []model.Commit

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := g.doRequest(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("github commits: %d %s", resp.StatusCode, string(b))
		}

		var raw []struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Name  string    `json:"name"`
					Email string    `json:"email"`
					Date  time.Time `json:"date"`
				} `json:"author"`
				Committer struct {
					Name  string    `json:"name"`
					Email string    `json:"email"`
					Date  time.Time `json:"date"`
				} `json:"committer"`
			} `json:"commit"`
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			Committer struct {
				Login string `json:"login"`
			} `json:"committer"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
			return nil, err
		}

		for _, r := range raw {
			committer := r.Committer.Login
			if committer == "" {
				committer = r.Commit.Committer.Name
			}
			author := r.Author.Login
			if author == "" {
				author = r.Commit.Author.Name
			}

			commits = append(commits, model.Commit{
				SHA:       r.SHA,
				Message:   r.Commit.Message,
				Committer: committer,
				Author:    author,
				Timestamp: r.Commit.Author.Date,
			})
		}

		url = getNextPage(resp)
	}

	return commits, nil
}

func (g *GithubClient) GetCommitDiff(owner, repo, sha string) (map[string]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", owner, repo, sha)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	resp, err := g.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github commit: %d %s", resp.StatusCode, string(b))
	}
	var raw struct {
		Files []struct {
			Filename string `json:"filename"`
			Patch    string `json:"patch"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, f := range raw.Files {
		out[f.Filename] = f.Patch
	}
	return out, nil
}
