package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/usecase"
)

// diffReadUseCase は DiffHandler が使うユースケースインターフェース。
// diff 本文のみ取得できればよく、グラフ読み出し（GetGraph）には依存しない。
type diffReadUseCase interface {
	GetDiff(ctx context.Context, id domain.JobID) (*domain.Analysis, error)
}

// DiffHandler は GET /api/diff/:jobId のハンドラ。
type DiffHandler struct {
	read diffReadUseCase
}

// NewDiffHandler は DiffHandler を返す。
func NewDiffHandler(read diffReadUseCase) *DiffHandler {
	return &DiffHandler{read: read}
}

// Get は GET /api/diff/:jobId を処理する。
func (h *DiffHandler) Get(c echo.Context) error {
	id := c.Param("jobId")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "jobId is required")
	}

	analysis, err := h.read.GetDiff(c.Request().Context(), domain.JobID(id))
	if err != nil {
		if errors.Is(err, usecase.ErrJobNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "job not found")
		}
		if errors.Is(err, usecase.ErrJobNotReady) {
			return echo.NewHTTPError(http.StatusConflict, "job not ready")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("get diff: %v", err))
	}

	c.Response().Header().Set("Cache-Control", snapshotCacheControl)
	return c.JSON(http.StatusOK, toDiffResponse(analysis))
}
