package github

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gogithub "github.com/google/go-github/v69/github"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// testUserRepo は httptest.Server の URL を向いた UserRepo を返すヘルパー。
func testUserRepo(baseURL string) *UserRepo {
	base, _ := url.Parse(baseURL + "/")
	return &UserRepo{
		newClient: func(ctx context.Context, _ string) *gogithub.Client {
			c := gogithub.NewClient(nil)
			c.BaseURL = base
			return c
		},
	}
}

func TestUserRepo_LoginByToken(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantLogin string
		wantErr   error // nil 以外なら errors.Is で判定
		anyErr    bool  // ErrInvalidToken 以外のエラーを期待する
	}{
		{name: "ok", status: http.StatusOK, body: `{"login":"octocat"}`, wantLogin: "octocat"},
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{"message":"Bad credentials"}`, wantErr: domain.ErrInvalidToken},
		{name: "empty login", status: http.StatusOK, body: `{}`, anyErr: true},
		{name: "not found", status: http.StatusNotFound, body: `{"message":"Not Found"}`, anyErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/user", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			got, err := testUserRepo(srv.URL).LoginByToken(context.Background(), "tok")
			switch {
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err: got %v, want %v", err, tt.wantErr)
				}
			case tt.anyErr:
				if err == nil || errors.Is(err, domain.ErrInvalidToken) {
					t.Fatalf("err: got %v, want non-ErrInvalidToken error", err)
				}
			default:
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.wantLogin {
					t.Errorf("login: got %q, want %q", got, tt.wantLogin)
				}
			}
		})
	}
}
