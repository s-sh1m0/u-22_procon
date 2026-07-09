package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/usecase"
)

func TestDiffGet_Success(t *testing.T) {
	analysis := &domain.Analysis{
		ID: "a1",
		ChangedFiles: []domain.DiffFile{
			{Filename: "pkg/a.go", Status: domain.FileStatusAdded, Additions: 3},
		},
	}
	h := NewDiffHandler(&fakeReadUC{analysis: analysis})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/diff/job-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("job-1")

	if err := h.Get(c); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	// jobId 単位で不変なスナップショットなので Cache-Control が付く。
	if got := rec.Header().Get("Cache-Control"); got != snapshotCacheControl {
		t.Errorf("Cache-Control=%q want %q", got, snapshotCacheControl)
	}
	var resp DiffResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Files) != 1 || resp.Files[0].Filename != "pkg/a.go" {
		t.Errorf("unexpected files: %+v", resp.Files)
	}
}

func TestDiffGet_NotReady(t *testing.T) {
	h := NewDiffHandler(&fakeReadUC{diffErr: usecase.ErrJobNotReady})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/diff/pending-job", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("pending-job")

	err := h.Get(c)
	if he, ok := err.(*echo.HTTPError); !ok || he.Code != http.StatusConflict {
		t.Errorf("expected 409, got %v", err)
	}
	// まだ done でない応答はキャッシュさせない。
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("Cache-Control should be unset on not-ready, got %q", got)
	}
}
