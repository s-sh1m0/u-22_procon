package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/github"
)

const (
	cookieSession    = "sid"
	cookieOAuthState = "oauth_state"
	sessionMaxAge    = 7 * 24 * 60 * 60 // 7 days
	oauthStateMaxAge = 5 * 60           // 5 minutes
)

// AuthHandler は /auth/* エンドポイントのハンドラ。
type AuthHandler struct {
	oauth       *github.OAuthConfig
	sessions    domain.SessionRepository
	frontendURL string
}

func NewAuthHandler(oauth *github.OAuthConfig, sessions domain.SessionRepository, frontendURL string) *AuthHandler {
	return &AuthHandler{oauth: oauth, sessions: sessions, frontendURL: frontendURL}
}

// Login は GET /auth/github。state を生成して Cookie にセットし、GitHub に Redirect する。
func (h *AuthHandler) Login(c echo.Context) error {
	state, err := randomToken(32)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "generate state")
	}
	c.SetCookie(&http.Cookie{
		Name:     cookieOAuthState,
		Value:    state,
		Path:     "/",
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   h.crossOrigin(),
		SameSite: http.SameSiteLaxMode,
	})
	return c.Redirect(http.StatusFound, h.oauth.AuthCodeURL(state))
}

// Callback は GET /auth/github/callback。state 検証 → token 交換 → セッション作成。
func (h *AuthHandler) Callback(c echo.Context) error {
	stateCookie, err := c.Cookie(cookieOAuthState)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "missing state cookie")
	}
	queryState := c.QueryParam("state")
	if subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(queryState)) != 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "state mismatch")
	}
	h.clearCookie(c, cookieOAuthState)

	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing code")
	}

	ctx := c.Request().Context()
	token, err := h.oauth.Exchange(ctx, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "exchange code: "+err.Error())
	}
	login, err := h.oauth.FetchUserLogin(ctx, token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "fetch user: "+err.Error())
	}

	sid, err := randomToken(32)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "generate sid")
	}
	session := &domain.Session{
		ID:          domain.SessionID(sid),
		GitHubToken: token.AccessToken,
		UserLogin:   login,
		CreatedAt:   time.Now().UTC(),
	}
	if err := h.sessions.Save(ctx, session); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "save session: "+err.Error())
	}

	c.SetCookie(&http.Cookie{
		Name:     cookieSession,
		Value:    sid,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   h.crossOrigin(),
		SameSite: h.sameSite(),
	})
	return c.Redirect(http.StatusFound, h.postLoginRedirect())
}

// Logout は POST /auth/logout。セッションを DB から削除し、Cookie を期限切れにする。
func (h *AuthHandler) Logout(c echo.Context) error {
	if cookie, err := c.Cookie(cookieSession); err == nil {
		_ = h.sessions.Delete(c.Request().Context(), domain.SessionID(cookie.Value))
	}
	h.clearCookie(c, cookieSession)
	return c.NoContent(http.StatusNoContent)
}

// Me は GET /auth/me。RequireAuth ミドルウェアが c.Set("session", ...) した値を返す。
func (h *AuthHandler) Me(c echo.Context) error {
	session, ok := c.Get(contextKeySession).(*domain.Session)
	if !ok || session == nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	return c.JSON(http.StatusOK, MeResponse{Login: session.UserLogin})
}

func randomToken(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (h *AuthHandler) crossOrigin() bool {
	return h.frontendURL != ""
}

func (h *AuthHandler) sameSite() http.SameSite {
	if h.crossOrigin() {
		return http.SameSiteNoneMode
	}
	return http.SameSiteLaxMode
}

func (h *AuthHandler) postLoginRedirect() string {
	if h.frontendURL != "" {
		return h.frontendURL + "/"
	}
	return "/"
}

func (h *AuthHandler) clearCookie(c echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.crossOrigin(),
		SameSite: h.sameSite(),
	})
}
