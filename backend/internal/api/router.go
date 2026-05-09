package api

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// NewRouter は Echo インスタンスを生成し全ルートを登録して返す。
// 認証保護が必要なエンドポイントには RequireAuth を適用する。
func NewRouter(
	authHandler *AuthHandler,
	analysisHandler *AnalysisHandler,
	jobHandler *JobHandler,
	diffHandler *DiffHandler,
	sessions domain.SessionRepository,
) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:  true,
		LogURI:     true,
		LogMethod:  true,
		LogLatency: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Printf("%s %s -> %d (%s)", v.Method, v.URI, v.Status, v.Latency)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	auth := e.Group("/auth")
	auth.GET("/github", authHandler.Login)
	auth.GET("/github/callback", authHandler.Callback)
	auth.POST("/logout", authHandler.Logout)
	auth.GET("/me", authHandler.Me, RequireAuth(sessions))

	apiG := e.Group("/api", RequireAuth(sessions))
	apiG.POST("/analyze", analysisHandler.Analyze)
	apiG.GET("/jobs/:id", jobHandler.Get)
	apiG.GET("/graph/:jobId", analysisHandler.GetGraph)
	apiG.GET("/diff/:jobId", diffHandler.Get)

	return e
}
