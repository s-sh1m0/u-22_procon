package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/usecase"
)

// --- fakes ---

type fakeAnalyzeUC struct {
	jobID domain.JobID
	err   error
}

func (f *fakeAnalyzeUC) Execute(_ context.Context, _ string, _ domain.PRInfo) (domain.JobID, error) {
	return f.jobID, f.err
}

type fakeReadUC struct {
	job      *domain.Job
	analysis *domain.Analysis
	jobErr   error
	graphErr error
}

func (f *fakeReadUC) GetJob(_ context.Context, _ domain.JobID) (*domain.Job, error) {
	return f.job, f.jobErr
}

func (f *fakeReadUC) GetGraph(_ context.Context, _ domain.JobID) (*domain.Analysis, error) {
	return f.analysis, f.graphErr
}

// --- helpers ---

func validSession() *domain.Session {
	return &domain.Session{ID: "sid", GitHubToken: "ghtok", UserLogin: "user", CreatedAt: time.Now()}
}

// --- parsePRURL tests ---

func TestParsePRURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		owner   string
		repo    string
		number  int
	}{
		{
			name:   "valid https",
			url:    "https://github.com/owner/repo/pull/123",
			owner:  "owner",
			repo:   "repo",
			number: 123,
		},
		{
			name:   "valid http",
			url:    "http://github.com/owner/repo/pull/1",
			owner:  "owner",
			repo:   "repo",
			number: 1,
		},
		{name: "wrong host", url: "https://gitlab.com/o/r/pull/1", wantErr: true},
		{name: "wrong scheme", url: "ftp://github.com/o/r/pull/1", wantErr: true},
		{name: "no number", url: "https://github.com/o/r/pull/", wantErr: true},
		{name: "not pull", url: "https://github.com/o/r/issues/1", wantErr: true},
		{name: "extra segment", url: "https://github.com/o/r/pull/1/files", wantErr: true},
		{name: "zero number", url: "https://github.com/o/r/pull/0", wantErr: true},
		{name: "negative number", url: "https://github.com/o/r/pull/-1", wantErr: true},
		{name: "empty", url: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePRURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil (result=%+v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Owner != tt.owner || got.Repo != tt.repo || got.Number != tt.number {
				t.Errorf("got %+v, want owner=%s repo=%s number=%d", got, tt.owner, tt.repo, tt.number)
			}
		})
	}
}

// --- Analyze endpoint tests ---

func TestAnalyze_Success(t *testing.T) {
	analyzeUC := &fakeAnalyzeUC{jobID: "job-1"}
	readUC := &fakeReadUC{job: &domain.Job{ID: "job-1", Status: domain.JobStatusPending}}
	h := NewAnalysisHandler(analyzeUC, readUC)

	e := echo.New()
	body := `{"pr_url":"https://github.com/owner/repo/pull/42"}`
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(contextKeySession, validSession())

	if err := h.Analyze(c); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp JobResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.JobID != "job-1" {
		t.Errorf("unexpected job_id: %s", resp.JobID)
	}
}

func TestAnalyze_InvalidURL(t *testing.T) {
	h := NewAnalysisHandler(&fakeAnalyzeUC{}, &fakeReadUC{})

	e := echo.New()
	body := `{"pr_url":"https://github.com/bad-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(contextKeySession, validSession())

	err := h.Analyze(c)
	if he, ok := err.(*echo.HTTPError); !ok || he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %v", err)
	}
}

func TestAnalyze_NoSession(t *testing.T) {
	h := NewAnalysisHandler(&fakeAnalyzeUC{}, &fakeReadUC{})

	e := echo.New()
	body := `{"pr_url":"https://github.com/owner/repo/pull/1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/analyze", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// no session set

	err := h.Analyze(c)
	if he, ok := err.(*echo.HTTPError); !ok || he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %v", err)
	}
}

// --- GetGraph endpoint tests ---

func TestGetGraph_Success(t *testing.T) {
	result := &domain.ClusterResult{
		Clusters: []domain.Cluster{{ID: 0, Label: "pkg", Nodes: []domain.NodeID{"fn:A"}}},
		Graph: domain.Graph{
			Nodes: []domain.Node{{ID: "fn:A", Name: "A", Changed: true}},
			Edges: []domain.Edge{},
		},
	}
	analysis := &domain.Analysis{ID: "a1", Result: result}
	h := NewAnalysisHandler(&fakeAnalyzeUC{}, &fakeReadUC{analysis: analysis})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/graph/job-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("job-1")

	if err := h.GetGraph(c); err != nil {
		t.Fatalf("GetGraph: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp GraphResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.Clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(resp.Clusters))
	}
	if len(resp.Graph.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(resp.Graph.Nodes))
	}
}

func TestGetGraph_NotFound(t *testing.T) {
	h := NewAnalysisHandler(&fakeAnalyzeUC{}, &fakeReadUC{graphErr: usecase.ErrJobNotFound})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/graph/missing", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("missing")

	err := h.GetGraph(c)
	if he, ok := err.(*echo.HTTPError); !ok || he.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %v", err)
	}
}

func TestGetGraph_NotReady(t *testing.T) {
	h := NewAnalysisHandler(&fakeAnalyzeUC{}, &fakeReadUC{graphErr: usecase.ErrJobNotReady})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/graph/pending-job", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("jobId")
	c.SetParamValues("pending-job")

	err := h.GetGraph(c)
	if he, ok := err.(*echo.HTTPError); !ok || he.Code != http.StatusConflict {
		t.Errorf("expected 409, got %v", err)
	}
}
