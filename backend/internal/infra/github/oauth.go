package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"
)

// OAuthConfig は GitHub OAuth2 の設定とユーザー情報取得をまとめた型。
type OAuthConfig struct {
	*oauth2.Config
}

// NewOAuthConfig は GitHub OAuth App の認証情報から OAuthConfig を生成する。
// scope は repo（プライベートリポジトリ含む PR diff 取得）と read:user。
func NewOAuthConfig(clientID, clientSecret, callbackURL string) *OAuthConfig {
	return &OAuthConfig{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  callbackURL,
			Scopes:       []string{"repo", "read:user"},
			Endpoint:     oauthgithub.Endpoint,
		},
	}
}

// FetchUserLogin は GitHub /user エンドポイントを叩いてログイン名のみを返す。
// go-github を使わずに済むよう最小限の実装に留める（PR 取得用クライアントは Issue #3 で追加）。
func (c *OAuthConfig) FetchUserLogin(ctx context.Context, token *oauth2.Token) (string, error) {
	client := c.Client(ctx, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return "", fmt.Errorf("build user request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github /user returned %d", resp.StatusCode)
	}
	var body struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode user: %w", err)
	}
	if body.Login == "" {
		return "", fmt.Errorf("github /user returned empty login")
	}
	return body.Login, nil
}
