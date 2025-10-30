# Repo Scanner
A lightweight Go web app that scans a GitHub repository for leaked AWS secrets by analyzing commit diffs across all branches. 
Supports resuming from the last scanned commit.

## Features
- Scans all branches via GitHub API
- Detects AWS access keys and secret-like values
- Supports resume from `from_sha`
- In-memory store for findings
- Simple REST API

## Setup
```bash
# Running from IDE
git clone https://github.com/yourusername/repo-secrets-scanner.git
cd repo-secrets-scanner
export GITHUB_TOKEN=your_token_here
export HTTP_ADDR=:8080
go run ./cmd/main.go

#
 Running via docker:
 
 docker build -t repo-secrets-scanner .
 
 docker run -d -p 8080:8080 \                           
  -e GITHUB_TOKEN=[TOKEN] \
  -e HTTP_ADDR=:8080 \
  repo-secrets-scanner

 
```

## API

(use localhost:8080)

POST /scan
Content-Type: application/json

```json
{
"owner": "repo-owner",
"repo": "repo-name",
"from_sha": "optional-last-processed-sha"
}
```


GET /findings?owner=[owner]&repo=[repo]

Response:
```json
[
{
  "commit_sha": "476f34937f74193547a2be347037e34c7724d1aa",
  "committer": "benami99",
  "file_path": "test_files/secret3.json",
  "leak_type": "aws_access_key",
  "leak_value": ".....",
  "timestamp": "2025-10-30T10:45:42Z"
}
]
```

GET /status?owner=[owner]&repo=[repo]

returns "running" or "done"


