package main

import (
	"log"
	"net/http"

	"github.com/benami99/repo-secrets-scanner/internal/api"
	"github.com/benami99/repo-secrets-scanner/internal/config"
	"github.com/benami99/repo-secrets-scanner/internal/repository"
	"github.com/benami99/repo-secrets-scanner/internal/service"
)

func main() {

	// Load config (port and token)
	cfg := config.Load()

	// create new memory store
	store := repository.NewMemoryStore()

	// create new github client
	gh := service.NewGithubClient(cfg.GithubToken)

	// create new scanner service with these storage and client
	scanner := service.NewScanner(store, gh)

	// API router
	r := api.NewRouter(scanner)

	// Start listen
	log.Printf("Listening on %s", cfg.HttpAddr)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
