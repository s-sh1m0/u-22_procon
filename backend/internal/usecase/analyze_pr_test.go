package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// --- fakes ---

type fakeAnalysisRepo struct {
	byID map[domain.AnalysisID]*domain.Analysis
	byPR *domain.Analysis // FindByPR が返す固定値
	err  error
}

func (r *fakeAnalysisRepo) Save(_ context.Context, a *domain.Analysis) error {
	if r.err != nil {
		return r.err
	}
	if r.byID == nil {
		r.byID = make(map[domain.AnalysisID]*domain.Analysis)
	}
	r.byID[a.ID] = a
	return nil
}

func (r *fakeAnalysisRepo) FindByID(_ context.Context, id domain.AnalysisID) (*domain.Analysis, error) {
	return r.byID[id], r.err
}

func (r *fakeAnalysisRepo) FindByPR(_ context.Context, _ domain.PRInfo) (*domain.Analysis, error) {
	return r.byPR, r.err
}

type fakeJobRepo struct {
	saved []*domain.Job
	err   error
}

func (r *fakeJobRepo) Save(_ context.Context, j *domain.Job) error {
	if r.err != nil {
		return r.err
	}
	r.saved = append(r.saved, j)
	return nil
}

func (r *fakeJobRepo) UpdateStatus(_ context.Context, _ domain.JobID, _ domain.JobStatus, _ string) error {
	return r.err
}

func (r *fakeJobRepo) FindByID(_ context.Context, id domain.JobID) (*domain.Job, error) {
	for _, j := range r.saved {
		if j.ID == id {
			return j, nil
		}
	}
	return nil, r.err
}

type fakeEnqueuer struct {
	calls []struct {
		jobID domain.JobID
		token string
	}
}

func (e *fakeEnqueuer) Enqueue(jobID domain.JobID, token string) {
	e.calls = append(e.calls, struct {
		jobID domain.JobID
		token string
	}{jobID, token})
}

// --- tests ---

func TestAnalyzePR_CacheMiss_Enqueue(t *testing.T) {
	analyses := &fakeAnalysisRepo{byPR: nil}
	jobs := &fakeJobRepo{}
	enqueuer := &fakeEnqueuer{}

	uc := NewAnalyzePRUseCase(analyses, jobs, enqueuer)
	pr := domain.PRInfo{Owner: "o", Repo: "r", Number: 1}

	jobID, err := uc.Execute(context.Background(), "token-xxx", pr)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if jobID == "" {
		t.Error("expected non-empty jobID")
	}
	if len(jobs.saved) != 1 {
		t.Fatalf("expected 1 saved job, got %d", len(jobs.saved))
	}
	if jobs.saved[0].Status != domain.JobStatusPending {
		t.Errorf("expected pending, got %s", jobs.saved[0].Status)
	}
	if len(enqueuer.calls) != 1 {
		t.Fatalf("expected 1 enqueue call, got %d", len(enqueuer.calls))
	}
	if enqueuer.calls[0].token != "token-xxx" {
		t.Errorf("unexpected token: %s", enqueuer.calls[0].token)
	}
}

func TestAnalyzePR_CacheHit_NoEnqueue(t *testing.T) {
	cachedAnalysis := &domain.Analysis{
		ID:           "cached-analysis",
		PR:           domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
		ChangedFiles: []domain.DiffFile{},
	}
	analyses := &fakeAnalysisRepo{byPR: cachedAnalysis}
	jobs := &fakeJobRepo{}
	enqueuer := &fakeEnqueuer{}

	uc := NewAnalyzePRUseCase(analyses, jobs, enqueuer)
	pr := domain.PRInfo{Owner: "o", Repo: "r", Number: 1}

	jobID, err := uc.Execute(context.Background(), "token-xxx", pr)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if jobID == "" {
		t.Error("expected non-empty jobID")
	}
	if len(jobs.saved) != 1 {
		t.Fatalf("expected 1 saved job, got %d", len(jobs.saved))
	}
	if jobs.saved[0].Status != domain.JobStatusDone {
		t.Errorf("expected done, got %s", jobs.saved[0].Status)
	}
	if jobs.saved[0].AnalysisID != "cached-analysis" {
		t.Errorf("expected cached-analysis, got %s", jobs.saved[0].AnalysisID)
	}
	if len(enqueuer.calls) != 0 {
		t.Errorf("expected no enqueue calls, got %d", len(enqueuer.calls))
	}
}

func TestAnalyzePR_RepoError_Propagated(t *testing.T) {
	sentinel := errors.New("db down")
	analyses := &fakeAnalysisRepo{err: sentinel}
	jobs := &fakeJobRepo{}
	enqueuer := &fakeEnqueuer{}

	uc := NewAnalyzePRUseCase(analyses, jobs, enqueuer)
	_, err := uc.Execute(context.Background(), "t", domain.PRInfo{Owner: "o", Repo: "r", Number: 1})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestAnalyzePR_IDsAreUnique(t *testing.T) {
	analyses := &fakeAnalysisRepo{}
	jobs := &fakeJobRepo{}
	enqueuer := &fakeEnqueuer{}

	uc := NewAnalyzePRUseCase(analyses, jobs, enqueuer)
	uc.now = func() time.Time { return time.Now() }

	pr := domain.PRInfo{Owner: "o", Repo: "r", Number: 1}
	id1, _ := uc.Execute(context.Background(), "t", pr)
	id2, _ := uc.Execute(context.Background(), "t", pr)
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}
