package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// JobHandler は GET /api/jobs/:id のハンドラ。
type JobHandler struct {
	read readUseCase
}

// NewJobHandler は JobHandler を返す。
func NewJobHandler(read readUseCase) *JobHandler {
	return &JobHandler{read: read}
}

// Get は GET /api/jobs/:id を処理する。
func (h *JobHandler) Get(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}

	j, err := h.read.GetJob(c.Request().Context(), domain.JobID(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("get job: %v", err))
	}
	if j == nil {
		return echo.NewHTTPError(http.StatusNotFound, "job not found")
	}

	return c.JSON(http.StatusOK, JobResponse{
		JobID:      string(j.ID),
		Status:     string(j.Status),
		Phase:      string(j.Phase),
		AnalysisID: string(j.AnalysisID),
		Error:      j.Error,
	})
}
