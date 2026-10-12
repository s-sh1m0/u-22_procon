package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	gogithub "github.com/google/go-github/v69/github"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// UserRepo は domain.GitHubUserRepository を go-github で実装する。
// newClient フィールドはテスト時に httptest 用クライアントへ差し替えられる。
type UserRepo struct {
	newClient func(ctx context.Context, token string) *gogithub.Client
}

// NewUserRepo は本番用の UserRepo を返す。
// レート制限対応・クライアントキャッシュ付きの ClientFactory を使う。
func NewUserRepo() *UserRepo {
	return &UserRepo{newClient: NewClientFactory()}
}

// LoginByToken は GitHub GET /user でトークンの持ち主のログイン名を返す。
// GitHub が 401 を返した場合（失効・偽造トークン）は domain.ErrInvalidToken を返す。
func (r *UserRepo) LoginByToken(ctx context.Context, token string) (string, error) {
	user, resp, err := r.newClient(ctx, token).Users.Get(ctx, "")
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized {
			return "", domain.ErrInvalidToken
		}
		return "", fmt.Errorf("github: get authenticated user: %w", err)
	}
	if user.GetLogin() == "" {
		return "", errors.New("github: /user returned empty login")
	}
	return user.GetLogin(), nil
}
