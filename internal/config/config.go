package config

import (
	"os"
)

type Config struct {
	GithubToken string
	HttpAddr    string
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func Load() Config {
	return Config{
		GithubToken: os.Getenv("GITHUB_TOKEN"),
		HttpAddr:    getEnv("HTTP_ADDR", ":8080"),
	}
}
