package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// fakeSessionRepo は ID → Session の固定マップを返す。
type fakeSessionRepo struct {
	byID map[domain.SessionID]*domain.Session
	err  error
}

func (r *fakeSessionRepo) Save(_ context.Context, _ *domain.Session) error { return r.err }

func (r *fakeSessionRepo) FindByID(_ context.Context, id domain.SessionID) (*domain.Session, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.byID[id], nil
}

func (r *fakeSessionRepo) Delete(_ context.Context, _ domain.SessionID) error { return r.err }

// fakeTokenAuthUC は token → login の固定マップで認証し、受け取ったトークンを記録する。
type fakeTokenAuthUC struct {
	logins map[string]string
	err    error
	got    []string
}

func (f *fakeTokenAuthUC) Execute(_ context.Context, token string) (*domain.Session, error) {
	f.got = append(f.got, token)
	if f.err != nil {
		return nil, f.err
	}
	login, ok := f.logins[token]
	if !ok {
		return nil, domain.ErrInvalidToken
	}
	return &domain.Session{GitHubToken: token, UserLogin: login}, nil
}

func TestRequireAuth(t *testing.T) {
	cookieSessions := &fakeSessionRepo{byID: map[domain.SessionID]*domain.Session{
		"sid-1": {ID: "sid-1", GitHubToken: "cookie-tok", UserLogin: "alice"},
	}}

	tests := []struct {
		name       string
		authHeader string
		cookie     string
		sessions   *fakeSessionRepo
		tokens     *fakeTokenAuthUC
		wantStatus int    // 0 なら next が呼ばれる
		wantLogin  string // next に渡ったセッションのログイン名
		wantToken  string // next に渡ったセッションの GitHub トークン
		wantTokens []string
	}{
		{
			name:       "有効な Bearer",
			authHeader: "Bearer gho_valid",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{logins: map[string]string{"gho_valid": "octocat"}},
			wantLogin:  "octocat",
			wantToken:  "gho_valid",
			wantTokens: []string{"gho_valid"},
		},
		{
			name:       "スキーム名は大文字小文字を区別しない",
			authHeader: "bearer gho_valid",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{logins: map[string]string{"gho_valid": "octocat"}},
			wantLogin:  "octocat",
			wantToken:  "gho_valid",
			wantTokens: []string{"gho_valid"},
		},
		{
			name:       "無効な Bearer は 401",
			authHeader: "Bearer gho_bad",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{logins: map[string]string{}},
			wantStatus: http.StatusUnauthorized,
			wantTokens: []string{"gho_bad"},
		},
		{
			name:       "Bearer が無効なら Cookie があっても 401",
			authHeader: "Bearer gho_bad",
			cookie:     "sid-1",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{logins: map[string]string{}},
			wantStatus: http.StatusUnauthorized,
			wantTokens: []string{"gho_bad"},
		},
		{
			name:       "GitHub 検証の失敗は 502",
			authHeader: "Bearer gho_valid",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{err: errors.New("github down")},
			wantStatus: http.StatusBadGateway,
			wantTokens: []string{"gho_valid"},
		},
		{
			name:       "Cookie の既存経路",
			cookie:     "sid-1",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{},
			wantLogin:  "alice",
			wantToken:  "cookie-tok",
			wantTokens: nil,
		},
		{
			name:       "Bearer 以外の Authorization は無視して Cookie を見る",
			authHeader: "Basic dXNlcjpwYXNz",
			cookie:     "sid-1",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{},
			wantLogin:  "alice",
			wantToken:  "cookie-tok",
			wantTokens: nil,
		},
		{
			name:       "未知の Cookie は 401",
			cookie:     "sid-unknown",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "セッション DB のエラーは 500",
			cookie:     "sid-1",
			sessions:   &fakeSessionRepo{err: errors.New("db down")},
			tokens:     &fakeTokenAuthUC{},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "どちらもなければ 401",
			sessions:   cookieSessions,
			tokens:     &fakeTokenAuthUC{},
			wantStatus: http.StatusUnauthorized,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/jobs/x", nil)
			if tt.authHeader != "" {
				req.Header.Set(echo.HeaderAuthorization, tt.authHeader)
			}
			if tt.cookie != "" {
				req.AddCookie(&http.Cookie{Name: cookieSession, Value: tt.cookie})
			}
			c := e.NewContext(req, httptest.NewRecorder())

			var gotSession *domain.Session
			next := func(c echo.Context) error {
				gotSession, _ = c.Get(contextKeySession).(*domain.Session)
				return nil
			}
			err := RequireAuth(tt.sessions, tt.tokens)(next)(c)

			if tt.wantStatus != 0 {
				var he *echo.HTTPError
				if !errors.As(err, &he) || he.Code != tt.wantStatus {
					t.Fatalf("err: got %v, want HTTP %d", err, tt.wantStatus)
				}
				if gotSession != nil {
					t.Errorf("next should not be called, got session %+v", gotSession)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if gotSession == nil {
					t.Fatal("next was not called with a session")
				}
				if gotSession.UserLogin != tt.wantLogin || gotSession.GitHubToken != tt.wantToken {
					t.Errorf("session: got login=%q token=%q, want login=%q token=%q",
						gotSession.UserLogin, gotSession.GitHubToken, tt.wantLogin, tt.wantToken)
				}
			}
			if len(tt.tokens.got) != len(tt.wantTokens) {
				t.Fatalf("token auth calls: got %v, want %v", tt.tokens.got, tt.wantTokens)
			}
			for i := range tt.wantTokens {
				if tt.tokens.got[i] != tt.wantTokens[i] {
					t.Errorf("token auth call %d: got %q, want %q", i, tt.tokens.got[i], tt.wantTokens[i])
				}
			}
		})
	}
}
