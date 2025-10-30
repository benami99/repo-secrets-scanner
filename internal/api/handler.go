package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/benami99/repo-secrets-scanner/internal/service"
	"github.com/go-chi/chi/v5"
)

func NewRouter(scanner *service.Scanner) http.Handler {
	r := chi.NewRouter()

	// Trigger a scan
	r.Post("/scan", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Owner   string `json:"owner"`
			Repo    string `json:"repo"`
			FromSHA string `json:"from_sha,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Stop here if a scan is already running
		if service.GetJobState(req.Owner, req.Repo) == "running" {
			w.WriteHeader(http.StatusConflict)
			_, err := fmt.Fprintf(w, "scan already running")
			if err != nil {
				return
			}
			return
		}

		// mark running
		service.SetJobState(req.Owner, req.Repo, "running")

		// run scan in goroutine
		go func() {
			defer service.SetJobState(req.Owner, req.Repo, "done")
			if err := scanner.ScanRepo(context.Background(), req.Owner, req.Repo, req.FromSHA); err != nil {
				fmt.Printf("scan failed: %v\n", err)
			}
		}()

		w.WriteHeader(http.StatusAccepted)
		_, err := fmt.Fprintf(w, "scan started")
		if err != nil {
			return
		}
	})

	// Get the status of the current scan
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		owner := r.URL.Query().Get("owner")
		repo := r.URL.Query().Get("repo")
		if owner == "" || repo == "" {
			http.Error(w, "owner and repo required", http.StatusBadRequest)
			return
		}
		state := service.GetJobState(owner, repo)
		_, err := fmt.Fprintf(w, "%s", state)
		if err != nil {
			return
		}
	})

	// List findings
	r.Get("/findings", func(w http.ResponseWriter, r *http.Request) {
		owner := r.URL.Query().Get("owner")
		repo := r.URL.Query().Get("repo")
		if owner == "" || repo == "" {
			http.Error(w, "owner and repo required", http.StatusBadRequest)
			return
		}

		// Use the scanner's ListFindings method
		findings := scanner.ListFindings(owner, repo)

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(findings)
		if err != nil {
			return
		}
	})

	return r
}
