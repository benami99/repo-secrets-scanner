package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// doRequest - handles rate limit
func (g *GithubClient) doRequest(req *http.Request) (*http.Response, error) {
	if g.token != "" {
		req.Header.Set("Authorization", "token "+g.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	for {
		resp, err := g.client.Do(req)
		if err != nil {
			return nil, err
		}

		// Handle rate limit
		if resp.StatusCode == 403 && resp.Header.Get("X-RateLimit-Remaining") == "0" {
			reset := resp.Header.Get("X-RateLimit-Reset")
			if reset != "" {
				resetUnix, _ := strconv.ParseInt(reset, 10, 64)
				wait := time.Until(time.Unix(resetUnix, 0))
				if wait > 0 {
					fmt.Printf("GitHub rate limit reached. Waiting %s...\n", wait)
					time.Sleep(wait)
					continue
				}
			}
		}

		return resp, nil
	}
}

// getNextPage - handles pagination
func getNextPage(resp *http.Response) string {
	link := resp.Header.Get("Link")
	if link == "" {
		return ""
	}

	parts := strings.Split(link, ",")
	for _, p := range parts {
		if strings.Contains(p, `rel="next"`) {
			start := strings.Index(p, "<")
			end := strings.Index(p, ">")
			if start != -1 && end != -1 {
				return p[start+1 : end]
			}
		}
	}
	return ""
}
