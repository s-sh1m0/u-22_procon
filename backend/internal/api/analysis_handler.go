package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/usecase"
)

// snapshotCacheControl は jobId 単位で不変な解析結果（グラフ / diff）に付ける
// Cache-Control。jobId は特定コミットのスナップショットで内容が変わらないため、
// クライアント側で長めにキャッシュしてよい。private はユーザー個別の認証済み応答のため。
const snapshotCacheControl = "private, max-age=86400"

// analyzeUseCase は AnalysisHandler が使うユースケースインターフェース。
type analyzeUseCase interface {
	Execute(ctx context.Context, token string, pr domain.PRInfo) (domain.JobID, error)
}

// readUseCase は AnalysisHandler と JobHandler が使うユースケースインターフェース。
type readUseCase interface {
	GetJob(ctx context.Context, id domain.JobID) (*domain.Job, error)
	GetGraph(ctx context.Context, id domain.JobID, mode domain.ClusterMode) (*domain.Analysis, error)
}

// AnalysisHandler は /api/analyze と /api/graph/:jobId のハンドラ。
type AnalysisHandler struct {
	analyze analyzeUseCase
	read    readUseCase
}

// NewAnalysisHandler は AnalysisHandler を返す。
func NewAnalysisHandler(analyze analyzeUseCase, read readUseCase) *AnalysisHandler {
	return &AnalysisHandler{analyze: analyze, read: read}
}

// Analyze は POST /api/analyze を処理する。
func (h *AnalysisHandler) Analyze(c echo.Context) error {
	var req AnalyzeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.PRURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "pr_url is required")
	}

	pr, err := parsePRURL(req.PRURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	sess, ok := c.Get(contextKeySession).(*domain.Session)
	if !ok || sess == nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	ctx := c.Request().Context()
	jobID, err := h.analyze.Execute(ctx, sess.GitHubToken, *pr)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("start analysis: %v", err))
	}

	j, err := h.read.GetJob(ctx, jobID)
	if err != nil || j == nil {
		return c.JSON(http.StatusOK, JobResponse{JobID: string(jobID), Status: "pending"})
	}
	return c.JSON(http.StatusOK, JobResponse{
		JobID:      string(j.ID),
		Status:     string(j.Status),
		AnalysisID: string(j.AnalysisID),
		Error:      j.Error,
	})
}

// GetGraph は GET /api/graph/:jobId を処理する。
// クエリパラメータ ?cluster=louvain|package|file でクラスタリング戦略を切り替える。
func (h *AnalysisHandler) GetGraph(c echo.Context) error {
	id := c.Param("jobId")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "jobId is required")
	}

	mode, err := domain.ParseClusterMode(c.QueryParam("cluster"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	analysis, err := h.read.GetGraph(c.Request().Context(), domain.JobID(id), mode)
	if err != nil {
		if errors.Is(err, usecase.ErrJobNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "job not found")
		}
		if errors.Is(err, usecase.ErrJobNotReady) {
			return echo.NewHTTPError(http.StatusConflict, "job not ready")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("get graph: %v", err))
	}

	c.Response().Header().Set("Cache-Control", snapshotCacheControl)
	return c.JSON(http.StatusOK, toGraphResponse(analysis))
}

// parsePRURL は GitHub PR の URL をパースして PRInfo を返す。
// 受け入れる形式: https://github.com/{owner}/{repo}/pull/{number}
func parsePRURL(rawURL string) (*domain.PRInfo, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("URL scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host != "github.com" {
		return nil, fmt.Errorf("URL host must be github.com, got %q", u.Host)
	}

	// path: /owner/repo/pull/number
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 {
		return nil, fmt.Errorf("URL path must be /owner/repo/pull/number, got %q", u.Path)
	}
	owner, repo, pullLiteral, numberStr := parts[0], parts[1], parts[2], parts[3]
	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo must not be empty")
	}
	if pullLiteral != "pull" {
		return nil, fmt.Errorf("URL must contain /pull/, got %q", pullLiteral)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil || number <= 0 {
		return nil, fmt.Errorf("PR number must be a positive integer, got %q", numberStr)
	}

	return &domain.PRInfo{Owner: owner, Repo: repo, Number: number}, nil
}
