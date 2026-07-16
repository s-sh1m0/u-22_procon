package github

import (
	"context"
	"net/http"
	"sync"

	gogithub "github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

const maxCachedClients = 64

// NewClientFactory はトークンごとにクライアントをキャッシュし、
// レート制限対応の go-github クライアントを返す関数を生成する。
// 同一トークンのリクエスト間でレート制限状態を共有する。
func NewClientFactory() func(ctx context.Context, accessToken string) *gogithub.Client {
	var mu sync.Mutex
	cache := make(map[string]*gogithub.Client)

	return func(_ context.Context, accessToken string) *gogithub.Client {
		if accessToken == "" {
			return gogithub.NewClient(nil)
		}

		mu.Lock()
		defer mu.Unlock()

		if c, ok := cache[accessToken]; ok {
			return c
		}

		if len(cache) >= maxCachedClients {
			clear(cache)
		}

		rlt := newRateLimitTransport(http.DefaultTransport)
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
		httpClient := &http.Client{
			Transport: &oauth2.Transport{Source: ts, Base: rlt},
		}
		c := gogithub.NewClient(httpClient)
		cache[accessToken] = c
		return c
	}
}
