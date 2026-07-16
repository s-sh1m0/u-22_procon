package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	frontendURL string,
	staticDir string,
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
	// レスポンス gzip 圧縮。グラフ / diff の JSON は ID の反復が多く圧縮率が高い。
	// MinLength 未満（ジョブ状態ポーリング等の小さな応答）は圧縮せず CPU を無駄にしない。
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{MinLength: 1024}))

	if frontendURL != "" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{frontendURL},
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
			AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAccept},
			AllowCredentials: true,
		}))
	}

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

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

	if staticDir != "" {
		e.GET("/*", func(c echo.Context) error {
			p := filepath.Join(staticDir, c.Request().URL.Path)
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return c.File(p)
			}
			return c.File(filepath.Join(staticDir, "index.html"))
		})
	}

	return e
}
