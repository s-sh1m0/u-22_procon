package api

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// contextKeySession は echo.Context に認証済みセッションを格納するキー。
const contextKeySession = "session"

// tokenAuthUseCase は Bearer トークン認証のユースケースインターフェース。
// トークンが無効な場合は domain.ErrInvalidToken を返す。
type tokenAuthUseCase interface {
	Execute(ctx context.Context, token string) (*domain.Session, error)
}

// RequireAuth は認証済みセッションを c.Set(contextKeySession, *domain.Session) してから
// 次のハンドラに進める。未認証なら 401 を返す。
//
// Authorization: Bearer <GitHub トークン> があればトークン認証（VS Code 拡張用）、
// なければ sid Cookie からセッションを引き当てる（Web 用）。
func RequireAuth(sessions domain.SessionRepository, tokens tokenAuthUseCase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if token, ok := bearerToken(c.Request()); ok {
				session, err := tokens.Execute(c.Request().Context(), token)
				if errors.Is(err, domain.ErrInvalidToken) {
					return echo.NewHTTPError(http.StatusUnauthorized)
				}
				if err != nil {
					log.Printf("auth: verify bearer token: %v", err)
					return echo.NewHTTPError(http.StatusBadGateway, "failed to verify GitHub token")
				}
				c.Set(contextKeySession, session)
				return next(c)
			}

			cookie, err := c.Cookie(cookieSession)
			if err != nil || cookie.Value == "" {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			session, err := sessions.FindByID(c.Request().Context(), domain.SessionID(cookie.Value))
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			if session == nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			c.Set(contextKeySession, session)
			return next(c)
		}
	}
}

// bearerToken は Authorization ヘッダが Bearer スキームならトークン部分と true を返す。
// スキーム名は大文字小文字を区別しない（RFC 7235）。トークンが空でも ok=true を返し、
// 検証側で無効トークンとして弾く。
func bearerToken(r *http.Request) (string, bool) {
	scheme, token, found := strings.Cut(r.Header.Get(echo.HeaderAuthorization), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	return strings.TrimSpace(token), true
}
