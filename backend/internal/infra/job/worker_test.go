package job

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/analyzer"
)

// --- fakes ---

type fakeJobStore struct {
	jobs        map[domain.JobID]*domain.Job
	statusCalls []struct {
		id     domain.JobID
		status domain.JobStatus
		errMsg string
	}
	analysisIDCalls []struct {
		id         domain.JobID
		analysisID domain.AnalysisID
	}
	phaseCalls []domain.JobPhase
	findErr    error
}

func newFakeJobStore() *fakeJobStore {
	return &fakeJobStore{jobs: make(map[domain.JobID]*domain.Job)}
}

func (s *fakeJobStore) FindByID(_ context.Context, id domain.JobID) (*domain.Job, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	return s.jobs[id], nil
}

func (s *fakeJobStore) UpdateStatus(_ context.Context, id domain.JobID, status domain.JobStatus, errMsg string) error {
	s.statusCalls = append(s.statusCalls, struct {
		id     domain.JobID
		status domain.JobStatus
		errMsg string
	}{id, status, errMsg})
	if j, ok := s.jobs[id]; ok {
		j.Status = status
	}
	return nil
}

func (s *fakeJobStore) UpdatePhase(_ context.Context, id domain.JobID, phase domain.JobPhase) error {
	s.phaseCalls = append(s.phaseCalls, phase)
	if j, ok := s.jobs[id]; ok {
		j.Phase = phase
	}
	return nil
}

func (s *fakeJobStore) UpdateAnalysisID(_ context.Context, id domain.JobID, analysisID domain.AnalysisID) error {
	s.analysisIDCalls = append(s.analysisIDCalls, struct {
		id         domain.JobID
		analysisID domain.AnalysisID
	}{id, analysisID})
	if j, ok := s.jobs[id]; ok {
		j.AnalysisID = analysisID
	}
	return nil
}

type fakeAnalysisRepo struct {
	saved []*domain.Analysis
}

func (r *fakeAnalysisRepo) Save(_ context.Context, a *domain.Analysis) error {
	r.saved = append(r.saved, a)
	return nil
}

func (r *fakeAnalysisRepo) FindByID(_ context.Context, _ domain.AnalysisID) (*domain.Analysis, error) {
	return nil, nil
}

func (r *fakeAnalysisRepo) FindByPR(_ context.Context, _ domain.PRInfo) (*domain.Analysis, error) {
	return nil, nil
}

type fakePRRepo struct {
	prInfo  *domain.PRInfo
	changed []domain.ChangedFile
	err     error
}

func (r *fakePRRepo) GetPR(_ context.Context, _, _, _ string, _ int) (*domain.PRInfo, error) {
	return r.prInfo, r.err
}

func (r *fakePRRepo) ListChangedGoFiles(_ context.Context, _, _, _ string, _ int) ([]domain.ChangedFile, error) {
	return r.changed, r.err
}

func (r *fakePRRepo) GetFileContent(_ context.Context, _, _, _, _, _ string) ([]byte, error) {
	return nil, r.err
}

type fakeSourceTree struct {
	prepared *analyzer.PreparedSource
	err      error

	// 観測用: 何回呼ばれたか / どの SHA で呼ばれたか
	prepareCalls      int
	prepareAtSHACalls int
	prepareAtSHAs     []string
	baseChanged       []domain.ChangedFile
}

func (s *fakeSourceTree) Prepare(_ context.Context, _ domain.PRInfo, _ string, _ []domain.ChangedFile) (*analyzer.PreparedSource, error) {
	s.prepareCalls++
	return s.prepared, s.err
}

func (s *fakeSourceTree) PrepareAtSHA(_ context.Context, _ domain.PRInfo, _, sha string, changed []domain.ChangedFile) (*analyzer.PreparedSource, error) {
	s.prepareAtSHACalls++
	s.prepareAtSHAs = append(s.prepareAtSHAs, sha)
	s.baseChanged = changed
	return s.prepared, s.err
}

type fakeCGBuilder struct {
	graph *domain.Graph
	err   error
}

func (b *fakeCGBuilder) Build(_ context.Context, _ []*packages.Package, _ []string, _ []string, _ string) (*domain.Graph, error) {
	return b.graph, b.err
}

type fakeClusterer struct {
	result *domain.ClusterResult
	err    error
}

func (c *fakeClusterer) Cluster(_ context.Context, _ *domain.Graph) (*domain.ClusterResult, error) {
	return c.result, c.err
}

// --- helpers ---

const fixedID = "fixed-id"

func newWorkerWithFakes(
	jobs *fakeJobStore,
	analyses *fakeAnalysisRepo,
	prRepo *fakePRRepo,
	sourceTree *fakeSourceTree,
	cgBuilder *fakeCGBuilder,
	clusterer *fakeClusterer,
) *Worker {
	q := NewQueue(1)
	w := NewWorker(q, jobs, analyses, prRepo, sourceTree, cgBuilder, clusterer)
	w.now = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }
	w.newID = func() string { return fixedID }
	return w
}

// --- tests ---

func TestWorker_HappyPath(t *testing.T) {
	prInfo := &domain.PRInfo{Owner: "o", Repo: "r", Number: 1, HeadSHA: "head-sha", BaseSHA: "base-sha"}
	graph := &domain.Graph{
		Nodes: []domain.Node{{ID: "fn:A", Name: "A"}},
	}
	result := &domain.ClusterResult{
		Clusters: []domain.Cluster{{ID: 0, Label: "pkg"}},
		Graph:    *graph,
	}

	jobs := newFakeJobStore()
	jobs.jobs["job-1"] = &domain.Job{
		ID:     "job-1",
		Status: domain.JobStatusPending,
		PR:     domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
	}
	analyses := &fakeAnalysisRepo{}

	st := &fakeSourceTree{prepared: &analyzer.PreparedSource{
		Packages:        nil,
		ChangedPackages: nil,
		Cleanup:         func() error { return nil },
	}}

	w := newWorkerWithFakes(
		jobs,
		analyses,
		&fakePRRepo{prInfo: prInfo, changed: []domain.ChangedFile{}},
		st,
		&fakeCGBuilder{graph: graph},
		&fakeClusterer{result: result},
	)

	w.process(context.Background(), Item{JobID: "job-1", Token: "tok"})

	// job should be done
	j := jobs.jobs["job-1"]
	if j.Status != domain.JobStatusDone {
		t.Errorf("expected done, got %s", j.Status)
	}
	if j.AnalysisID != fixedID {
		t.Errorf("expected fixed-id, got %s", j.AnalysisID)
	}
	// analysis should be saved
	if len(analyses.saved) != 1 {
		t.Fatalf("expected 1 analysis saved, got %d", len(analyses.saved))
	}
	if analyses.saved[0].ID != fixedID {
		t.Errorf("unexpected analysis ID: %s", analyses.saved[0].ID)
	}
	// フェーズが順番どおりに記録されたことを確認（build_graph は sync.Once で 1 回のみ）
	wantPhases := []domain.JobPhase{
		domain.JobPhaseClone,
		domain.JobPhaseBuildGraph,
		domain.JobPhaseDiff,
		domain.JobPhaseCluster,
		domain.JobPhaseVisualize,
	}
	if len(jobs.phaseCalls) != len(wantPhases) {
		t.Fatalf("phase calls: got %v want %v", jobs.phaseCalls, wantPhases)
	}
	for i, p := range wantPhases {
		if jobs.phaseCalls[i] != p {
			t.Errorf("phase[%d]: got %s want %s", i, jobs.phaseCalls[i], p)
		}
	}
	// base/head が両方呼ばれたことを確認
	if st.prepareCalls != 1 {
		t.Errorf("Prepare (head) calls: got %d want 1", st.prepareCalls)
	}
	if st.prepareAtSHACalls != 1 {
		t.Errorf("PrepareAtSHA (base) calls: got %d want 1", st.prepareAtSHACalls)
	}
	if len(st.prepareAtSHAs) != 1 || st.prepareAtSHAs[0] != "base-sha" {
		t.Errorf("PrepareAtSHA SHA: got %v want [base-sha]", st.prepareAtSHAs)
	}
}

func TestWorker_BaseSHA_RenameNormalized(t *testing.T) {
	prInfo := &domain.PRInfo{Owner: "o", Repo: "r", Number: 1, HeadSHA: "head", BaseSHA: "base"}
	graph := &domain.Graph{}
	result := &domain.ClusterResult{Graph: *graph}

	jobs := newFakeJobStore()
	jobs.jobs["job-r"] = &domain.Job{
		ID:     "job-r",
		Status: domain.JobStatusPending,
		PR:     domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
	}
	analyses := &fakeAnalysisRepo{}

	changed := []domain.ChangedFile{
		{Filename: "pkg/a/new.go", PreviousFilename: "pkg/a/old.go", Status: domain.FileStatusRenamed},
		{Filename: "pkg/b/added.go", Status: domain.FileStatusAdded},
		{Filename: "pkg/c/removed.go", Status: domain.FileStatusRemoved},
	}

	st := &fakeSourceTree{prepared: &analyzer.PreparedSource{Cleanup: func() error { return nil }}}

	w := newWorkerWithFakes(
		jobs, analyses,
		&fakePRRepo{prInfo: prInfo, changed: changed},
		st,
		&fakeCGBuilder{graph: graph},
		&fakeClusterer{result: result},
	)

	w.process(context.Background(), Item{JobID: "job-r", Token: "tok"})

	if len(st.baseChanged) != 2 {
		t.Fatalf("base changed len=%d want 2 (added 除外)", len(st.baseChanged))
	}
	// renamed が PreviousFilename に置換されているか
	var found bool
	for _, f := range st.baseChanged {
		if f.Status == domain.FileStatusRenamed && f.Filename == "pkg/a/old.go" {
			found = true
		}
		if f.Status == domain.FileStatusAdded {
			t.Error("added should be excluded for base")
		}
	}
	if !found {
		t.Error("renamed file should use PreviousFilename for base")
	}
}

func TestWorker_PRRepoError_MarksJobError(t *testing.T) {
	jobs := newFakeJobStore()
	jobs.jobs["job-2"] = &domain.Job{
		ID:     "job-2",
		Status: domain.JobStatusPending,
		PR:     domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
	}
	analyses := &fakeAnalysisRepo{}

	w := newWorkerWithFakes(
		jobs, analyses,
		&fakePRRepo{err: errors.New("github API error")},
		&fakeSourceTree{},
		&fakeCGBuilder{},
		&fakeClusterer{},
	)

	w.process(context.Background(), Item{JobID: "job-2", Token: "tok"})

	j := jobs.jobs["job-2"]
	if j.Status != domain.JobStatusError {
		t.Errorf("expected error status, got %s", j.Status)
	}
}

func TestWorker_ContextCancel_StopsLoop(t *testing.T) {
	q := NewQueue(1)
	jobs := newFakeJobStore()
	analyses := &fakeAnalysisRepo{}
	w := &Worker{
		queue:    q,
		jobs:     jobs,
		analyses: analyses,
		prRepo:   &fakePRRepo{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Run did not stop after context cancel")
	}
}

func TestWorker_QueueClose_StopsLoop(t *testing.T) {
	q := NewQueue(1)
	jobs := newFakeJobStore()
	analyses := &fakeAnalysisRepo{}
	w := &Worker{
		queue:    q,
		jobs:     jobs,
		analyses: analyses,
		prRepo:   &fakePRRepo{},
	}

	done := make(chan error, 1)
	go func() { done <- w.Run(context.Background()) }()

	q.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil on close, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Run did not stop after queue close")
	}
}
