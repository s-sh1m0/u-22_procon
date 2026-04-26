package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// NewRouter は Echo インスタンスを生成し全ルートを登録して返す。
// 認証保護が必要なエンドポイントには RequireAuth を適用する。
// /api/* ルートは Issue #3 以降で追加予定。
func NewRouter(authHandler *AuthHandler, sessions domain.SessionRepository) *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	auth := e.Group("/auth")
	auth.GET("/github", authHandler.Login)
	auth.GET("/github/callback", authHandler.Callback)
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/me", authHandler.Me, RequireAuth(sessions))

	return e
}
