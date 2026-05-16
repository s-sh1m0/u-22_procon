package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

type fakeGetJobRepo struct {
	jobs map[domain.JobID]*domain.Job
	err  error
}

func (r *fakeGetJobRepo) Save(_ context.Context, j *domain.Job) error {
	if r.jobs == nil {
		r.jobs = make(map[domain.JobID]*domain.Job)
	}
	r.jobs[j.ID] = j
	return nil
}

func (r *fakeGetJobRepo) UpdateStatus(_ context.Context, _ domain.JobID, _ domain.JobStatus, _ string) error {
	return r.err
}

func (r *fakeGetJobRepo) FindByID(_ context.Context, id domain.JobID) (*domain.Job, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.jobs[id], nil
}

type fakeGetAnalysisRepo struct {
	analyses map[domain.AnalysisID]*domain.Analysis
	err      error
}

func (r *fakeGetAnalysisRepo) Save(_ context.Context, a *domain.Analysis) error {
	if r.analyses == nil {
		r.analyses = make(map[domain.AnalysisID]*domain.Analysis)
	}
	r.analyses[a.ID] = a
	return nil
}

func (r *fakeGetAnalysisRepo) FindByID(_ context.Context, id domain.AnalysisID) (*domain.Analysis, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.analyses[id], nil
}

func (r *fakeGetAnalysisRepo) FindByPR(_ context.Context, _ domain.PRInfo) (*domain.Analysis, error) {
	return nil, r.err
}

func TestGetGraph_NotFound(t *testing.T) {
	jobs := &fakeGetJobRepo{jobs: make(map[domain.JobID]*domain.Job)}
	analyses := &fakeGetAnalysisRepo{analyses: make(map[domain.AnalysisID]*domain.Analysis)}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	_, err := uc.GetGraph(context.Background(), "nonexistent", domain.ClusterModeLouvain)
	if !errors.Is(err, ErrJobNotFound) {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}

func TestGetGraph_NotReady_Pending(t *testing.T) {
	jobs := &fakeGetJobRepo{jobs: map[domain.JobID]*domain.Job{
		"j1": {ID: "j1", Status: domain.JobStatusPending},
	}}
	analyses := &fakeGetAnalysisRepo{analyses: make(map[domain.AnalysisID]*domain.Analysis)}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	_, err := uc.GetGraph(context.Background(), "j1", domain.ClusterModeLouvain)
	if !errors.Is(err, ErrJobNotReady) {
		t.Errorf("expected ErrJobNotReady, got %v", err)
	}
}

func TestGetGraph_NotReady_DoneButNoAnalysisID(t *testing.T) {
	jobs := &fakeGetJobRepo{jobs: map[domain.JobID]*domain.Job{
		"j2": {ID: "j2", Status: domain.JobStatusDone, AnalysisID: ""},
	}}
	analyses := &fakeGetAnalysisRepo{analyses: make(map[domain.AnalysisID]*domain.Analysis)}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	_, err := uc.GetGraph(context.Background(), "j2", domain.ClusterModeLouvain)
	if !errors.Is(err, ErrJobNotReady) {
		t.Errorf("expected ErrJobNotReady, got %v", err)
	}
}

func TestGetGraph_Success(t *testing.T) {
	analysis := &domain.Analysis{
		ID:        "a1",
		PR:        domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
		Result:    &domain.ClusterResult{},
		CreatedAt: time.Now(),
	}
	jobs := &fakeGetJobRepo{jobs: map[domain.JobID]*domain.Job{
		"j3": {ID: "j3", Status: domain.JobStatusDone, AnalysisID: "a1"},
	}}
	analyses := &fakeGetAnalysisRepo{analyses: map[domain.AnalysisID]*domain.Analysis{
		"a1": analysis,
	}}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	got, err := uc.GetGraph(context.Background(), "j3", domain.ClusterModeLouvain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != "a1" {
		t.Errorf("unexpected analysis: %+v", got)
	}
}

// GetGraph はリポジトリから取得した Analysis を直接書き換えてはならない。
// （将来リポジトリがキャッシュ化された際に副作用が漏れないようにするための回帰テスト）
func TestGetGraph_DoesNotMutateRepoAnalysis(t *testing.T) {
	original := &domain.ClusterResult{
		Clusters: []domain.Cluster{{ID: 0, Label: "louvain-0", Nodes: []domain.NodeID{"fn:A"}}},
		Graph: domain.Graph{
			Nodes: []domain.Node{
				{ID: "fn:A", Package: "pkg/a", File: "a.go"},
				{ID: "fn:B", Package: "pkg/b", File: "b.go"},
			},
		},
	}
	stored := &domain.Analysis{ID: "a1", Result: original}
	jobs := &fakeGetJobRepo{jobs: map[domain.JobID]*domain.Job{
		"j1": {ID: "j1", Status: domain.JobStatusDone, AnalysisID: "a1"},
	}}
	analyses := &fakeGetAnalysisRepo{analyses: map[domain.AnalysisID]*domain.Analysis{
		"a1": stored,
	}}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	// package モードで取得（regroup により Result が差し替わる）
	if _, err := uc.GetGraph(context.Background(), "j1", domain.ClusterModePackage); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// リポジトリ内の Analysis.Result が書き換わっていないこと
	if stored.Result != original {
		t.Fatalf("repo analysis.Result was replaced; want same pointer as original")
	}
	if len(stored.Result.Clusters) != 1 || stored.Result.Clusters[0].Label != "louvain-0" {
		t.Errorf("original clusters mutated: %+v", stored.Result.Clusters)
	}

	// 続けて louvain で取得したら元の結果が返ること
	got, err := uc.GetGraph(context.Background(), "j1", domain.ClusterModeLouvain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result.Clusters) != 1 || got.Result.Clusters[0].Label != "louvain-0" {
		t.Errorf("louvain result was contaminated by previous package call: %+v", got.Result.Clusters)
	}
}

func TestGetGraph_IntegrityError(t *testing.T) {
	jobs := &fakeGetJobRepo{jobs: map[domain.JobID]*domain.Job{
		"j4": {ID: "j4", Status: domain.JobStatusDone, AnalysisID: "missing-analysis"},
	}}
	analyses := &fakeGetAnalysisRepo{analyses: make(map[domain.AnalysisID]*domain.Analysis)}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	_, err := uc.GetGraph(context.Background(), "j4", domain.ClusterModeLouvain)
	if err == nil {
		t.Error("expected error for missing analysis, got nil")
	}
}

func TestGetJob_Delegates(t *testing.T) {
	jobs := &fakeGetJobRepo{jobs: map[domain.JobID]*domain.Job{
		"j5": {ID: "j5", Status: domain.JobStatusRunning},
	}}
	analyses := &fakeGetAnalysisRepo{}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	got, err := uc.GetJob(context.Background(), "j5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != "j5" {
		t.Errorf("unexpected job: %+v", got)
	}
}

func TestGetJob_RepoError(t *testing.T) {
	sentinel := errors.New("db error")
	jobs := &fakeGetJobRepo{err: sentinel}
	analyses := &fakeGetAnalysisRepo{}
	uc := NewGetAnalysisUseCase(analyses, jobs)

	_, err := uc.GetJob(context.Background(), "any")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
