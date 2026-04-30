package github

import (
	"context"
	"net/http"

	gogithub "github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

// NewClient は accessToken でユーザー認証された go-github クライアントを返す。
// PRRepo 内からリクエストごとに呼ばれる想定。
func NewClient(ctx context.Context, accessToken string) *gogithub.Client {
	var httpClient *http.Client
	if accessToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
		httpClient = oauth2.NewClient(ctx, ts)
	}
	return gogithub.NewClient(httpClient)
}
