package service

import "regexp"

// -------------------------------------------------------
// Leak detection
// -------------------------------------------------------

type leak struct {
	Type  string
	Value string
}

var (
	reAwsKey    = regexp.MustCompile(`\b(AKIA|ASIA|ANPA|AROA|AIDA)[A-Z0-9]{16}\b`)
	reAwsSecret = regexp.MustCompile(`(?i)["']?[A-Za-z0-9/+=]{40}["']?`)
)

func detectLeaks(content string) []leak {
	out := []leak{}

	for _, m := range reAwsKey.FindAllString(content, -1) {
		out = append(out, leak{Type: "aws_access_key", Value: m})
	}
	for _, m := range reAwsSecret.FindAllString(content, -1) {
		if len(m) == 40 {
			out = append(out, leak{Type: "aws_secret_key", Value: m})
		}
	}
	return out
}
