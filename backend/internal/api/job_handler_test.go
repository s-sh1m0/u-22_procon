package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func TestJobHandler_Get_Success(t *testing.T) {
	job := &domain.Job{
		ID:         "job-1",
		Status:     domain.JobStatusDone,
		AnalysisID: "analysis-1",
	}
	h := NewJobHandler(&fakeReadUC{job: job})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/job-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("job-1")

	if err := h.Get(c); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp JobResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.JobID != "job-1" || resp.Status != "done" || resp.AnalysisID != "analysis-1" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestJobHandler_Get_NotFound(t *testing.T) {
	h := NewJobHandler(&fakeReadUC{job: nil})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("missing")

	err := h.Get(c)
	if he, ok := err.(*echo.HTTPError); !ok || he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %v", err)
	}
}

func TestJobHandler_Get_Pending(t *testing.T) {
	job := &domain.Job{ID: "job-2", Status: domain.JobStatusPending}
	h := NewJobHandler(&fakeReadUC{job: job})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/job-2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("job-2")

	if err := h.Get(c); err != nil {
		t.Fatalf("Get: %v", err)
	}
	var resp JobResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.AnalysisID != "" {
		t.Errorf("analysis_id should be omitted for pending, got %q", resp.AnalysisID)
	}
}
