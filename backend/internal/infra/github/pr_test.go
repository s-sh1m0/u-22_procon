package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gogithub "github.com/google/go-github/v69/github"
)

// testPRRepo は httptest.Server の URL を向いた PRRepo を返すヘルパー。
func testPRRepo(baseURL string) *PRRepo {
	base, _ := url.Parse(baseURL + "/")
	return &PRRepo{
		newClient: func(ctx context.Context, _ string) *gogithub.Client {
			c := gogithub.NewClient(nil)
			c.BaseURL = base
			return c
		},
	}
}

func TestPRRepo_GetPR(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"title": "fix: something",
			"base":  map[string]any{"ref": "main"},
			"head":  map[string]any{"ref": "feature", "sha": "abc123"},
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	repo := testPRRepo(srv.URL)
	got, err := repo.GetPR(context.Background(), "tok", "o", "r", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "fix: something" {
		t.Errorf("Title: got %q, want %q", got.Title, "fix: something")
	}
	if got.BaseRef != "main" {
		t.Errorf("BaseRef: got %q, want %q", got.BaseRef, "main")
	}
	if got.HeadRef != "feature" {
		t.Errorf("HeadRef: got %q, want %q", got.HeadRef, "feature")
	}
	if got.HeadSHA != "abc123" {
		t.Errorf("HeadSHA: got %q, want %q", got.HeadSHA, "abc123")
	}
	if got.Owner != "o" || got.Repo != "r" || got.Number != 1 {
		t.Errorf("PRInfo owner/repo/number: got %+v", got)
	}
}

func TestPRRepo_ListChangedGoFiles(t *testing.T) {
	rawFiles := []map[string]any{
		{
			"filename":  "main.go",
			"status":    "modified",
			"additions": 3,
			"deletions": 1,
			"patch":     "@@ -1,1 +1,3 @@\n-old\n+new\n+line2\n+line3",
		},
		{
			"filename":  "README.md",
			"status":    "modified",
			"additions": 1,
			"deletions": 0,
			"patch":     "@@ -1 +1,2 @@\n line\n+added",
		},
		{
			"filename":  "internal/handler.go",
			"status":    "added",
			"additions": 10,
			"deletions": 0,
			"patch":     "@@ -0,0 +1,10 @@\n+package main",
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/o/r/pulls/1/files", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rawFiles)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	repo := testPRRepo(srv.URL)
	got, err := repo.ListChangedGoFiles(context.Background(), "tok", "o", "r", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// README.md は除外されるので .go の2件のみ
	if len(got) != 2 {
		t.Fatalf("len: got %d, want 2", len(got))
	}

	// main.go の検証
	if got[0].Filename != "main.go" {
		t.Errorf("Filename: got %q, want %q", got[0].Filename, "main.go")
	}
	if got[0].Status != "modified" {
		t.Errorf("Status: got %q, want %q", got[0].Status, "modified")
	}
	if got[0].Additions != 3 || got[0].Deletions != 1 {
		t.Errorf("Additions/Deletions: got %d/%d", got[0].Additions, got[0].Deletions)
	}
	if got[0].Patch == "" {
		t.Error("Patch should not be empty")
	}

	// internal/handler.go の検証
	if got[1].Filename != "internal/handler.go" {
		t.Errorf("Filename: got %q, want %q", got[1].Filename, "internal/handler.go")
	}
}
