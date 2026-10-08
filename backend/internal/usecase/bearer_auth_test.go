package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// fakeUserRepo は token → login の対応を返し、呼び出し回数を数える。
type fakeUserRepo struct {
	logins map[string]string
	err    error
	calls  int
}

func (r *fakeUserRepo) LoginByToken(_ context.Context, token string) (string, error) {
	r.calls++
	if r.err != nil {
		return "", r.err
	}
	login, ok := r.logins[token]
	if !ok {
		return "", domain.ErrInvalidToken
	}
	return login, nil
}

func TestAuthenticateTokenUseCase_Execute(t *testing.T) {
	const ttl = 10 * time.Minute
	base := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	errGitHubDown := errors.New("github down")

	tests := []struct {
		name      string
		repo      *fakeUserRepo
		tokens    []string        // 順に Execute する
		advance   []time.Duration // 各 Execute 前に進める時間（tokens と同じ長さ）
		wantLogin string          // 最後の Execute の結果
		wantErr   error
		wantCalls int
	}{
		{
			name:      "初回は GitHub に問い合わせる",
			repo:      &fakeUserRepo{logins: map[string]string{"tok": "octocat"}},
			tokens:    []string{"tok"},
			advance:   []time.Duration{0},
			wantLogin: "octocat",
			wantCalls: 1,
		},
		{
			name:      "TTL 内の 2 回目はキャッシュを使う",
			repo:      &fakeUserRepo{logins: map[string]string{"tok": "octocat"}},
			tokens:    []string{"tok", "tok"},
			advance:   []time.Duration{0, ttl - time.Second},
			wantLogin: "octocat",
			wantCalls: 1,
		},
		{
			name:      "TTL 切れで再検証する",
			repo:      &fakeUserRepo{logins: map[string]string{"tok": "octocat"}},
			tokens:    []string{"tok", "tok"},
			advance:   []time.Duration{0, ttl},
			wantLogin: "octocat",
			wantCalls: 2,
		},
		{
			name:      "別トークンは別エントリ",
			repo:      &fakeUserRepo{logins: map[string]string{"a": "alice", "b": "bob"}},
			tokens:    []string{"a", "b"},
			advance:   []time.Duration{0, 0},
			wantLogin: "bob",
			wantCalls: 2,
		},
		{
			name:      "無効トークンは ErrInvalidToken",
			repo:      &fakeUserRepo{logins: map[string]string{}},
			tokens:    []string{"bad"},
			advance:   []time.Duration{0},
			wantErr:   domain.ErrInvalidToken,
			wantCalls: 1,
		},
		{
			name:      "無効トークンはキャッシュしない",
			repo:      &fakeUserRepo{logins: map[string]string{}},
			tokens:    []string{"bad", "bad"},
			advance:   []time.Duration{0, 0},
			wantErr:   domain.ErrInvalidToken,
			wantCalls: 2,
		},
		{
			name:      "GitHub 側のエラーはラップして返す",
			repo:      &fakeUserRepo{err: errGitHubDown},
			tokens:    []string{"tok"},
			advance:   []time.Duration{0},
			wantErr:   errGitHubDown,
			wantCalls: 1,
		},
		{
			name:      "空トークンは問い合わせずに ErrInvalidToken",
			repo:      &fakeUserRepo{},
			tokens:    []string{""},
			advance:   []time.Duration{0},
			wantErr:   domain.ErrInvalidToken,
			wantCalls: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewAuthenticateTokenUseCase(tt.repo, ttl)
			now := base
			uc.now = func() time.Time { return now }

			var (
				got *domain.Session
				err error
			)
			for i, tok := range tt.tokens {
				now = now.Add(tt.advance[i])
				got, err = uc.Execute(context.Background(), tok)
			}

			if tt.repo.calls != tt.wantCalls {
				t.Errorf("LoginByToken calls: got %d, want %d", tt.repo.calls, tt.wantCalls)
			}
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err: got %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			last := tt.tokens[len(tt.tokens)-1]
			if got.UserLogin != tt.wantLogin || got.GitHubToken != last || got.ID != "" {
				t.Errorf("session: got %+v, want login=%q token=%q id=\"\"", got, tt.wantLogin, last)
			}
		})
	}
}

func TestAuthenticateTokenUseCase_PrunesExpired(t *testing.T) {
	repo := &fakeUserRepo{logins: map[string]string{"a": "alice", "b": "bob"}}
	uc := NewAuthenticateTokenUseCase(repo, time.Minute)
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return now }

	if _, err := uc.Execute(context.Background(), "a"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := uc.Execute(context.Background(), "b"); err != nil {
		t.Fatal(err)
	}
	if len(uc.cache) != 1 {
		t.Errorf("cache size: got %d, want 1 (expired entry should be pruned)", len(uc.cache))
	}
}
