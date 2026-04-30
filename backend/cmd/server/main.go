package main

import (
	"log"
	"os"

	"github.com/s-sh1m0/u-22_procon/backend/internal/api"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/github"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/store"
)

func main() {
	clientID := mustEnv("GITHUB_CLIENT_ID")
	clientSecret := mustEnv("GITHUB_CLIENT_SECRET")
	callbackURL := mustEnv("GITHUB_CALLBACK_URL")
	sessionSecret := mustEnv("SESSION_SECRET")
	port := envOr("PORT", "8080")
	dbPath := envOr("DB_PATH", "/app/data/reviewarena.db")

	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer func() { _ = db.Close() }()

	sessions, err := store.NewSessionRepo(db, []byte(sessionSecret))
	if err != nil {
		log.Fatalf("init session repo: %v", err)
	}

	oauth := github.NewOAuthConfig(clientID, clientSecret, callbackURL)
	authHandler := api.NewAuthHandler(oauth, sessions)
	router := api.NewRouter(authHandler, sessions)

	router.Logger.Fatal(router.Start(":" + port))
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env %s is not set", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
