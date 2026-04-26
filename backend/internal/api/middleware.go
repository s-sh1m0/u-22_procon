package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// contextKeySession は echo.Context に認証済みセッションを格納するキー。
const contextKeySession = "session"

// RequireAuth は sid Cookie からセッションを引き当て、c.Set(contextKeySession, *domain.Session)
// してから次のハンドラに進める。未認証なら 401 を返す。
func RequireAuth(sessions domain.SessionRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
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
