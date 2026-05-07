package job

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/analyzer"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/cluster"
)

// jobStore はワーカーが使うジョブ永続化インターフェース。
type jobStore interface {
	FindByID(ctx context.Context, id domain.JobID) (*domain.Job, error)
	UpdateStatus(ctx context.Context, id domain.JobID, status domain.JobStatus, errMsg string) error
	UpdateAnalysisID(ctx context.Context, id domain.JobID, analysisID domain.AnalysisID) error
}

// sourceTreePreparer はワーカーが使うソースツリー準備インターフェース。
type sourceTreePreparer interface {
	Prepare(ctx context.Context, pr domain.PRInfo, token string, changed []domain.ChangedFile) (*analyzer.PreparedSource, error)
}

// Worker はキューからジョブを受け取り解析パイプラインを実行する。
type Worker struct {
	queue      *Queue
	jobs       jobStore
	analyses   domain.AnalysisRepository
	prRepo     domain.PRRepository
	sourceTree sourceTreePreparer
	cgBuilder  analyzer.CallGraphBuilder
	clusterer  cluster.Clusterer
	now        func() time.Time
	newID      func() string
}

// NewWorker は Worker を返す。
func NewWorker(
	queue *Queue,
	jobs jobStore,
	analyses domain.AnalysisRepository,
	prRepo domain.PRRepository,
	sourceTree sourceTreePreparer,
	cgBuilder analyzer.CallGraphBuilder,
	clusterer cluster.Clusterer,
) *Worker {
	return &Worker{
		queue:      queue,
		jobs:       jobs,
		analyses:   analyses,
		prRepo:     prRepo,
		sourceTree: sourceTree,
		cgBuilder:  cgBuilder,
		clusterer:  clusterer,
		now:        time.Now,
		newID:      workerRandomID,
	}
}

// Run はジョブキューを読み込み、ctx がキャンセルされるかキューが閉じるまでブロックする。
func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, ok := <-w.queue.Recv():
			if !ok {
				return nil
			}
			w.process(ctx, item)
		}
	}
}

func (w *Worker) process(ctx context.Context, item Item) {
	jobID := item.JobID
	fail := func(err error) {
		log.Printf("worker: job %s failed: %v", jobID, err)
		_ = w.jobs.UpdateStatus(ctx, jobID, domain.JobStatusError, err.Error())
	}

	if err := w.jobs.UpdateStatus(ctx, jobID, domain.JobStatusRunning, ""); err != nil {
		log.Printf("worker: failed to mark job %s running: %v", jobID, err)
	}

	j, err := w.jobs.FindByID(ctx, jobID)
	if err != nil {
		fail(fmt.Errorf("find job: %w", err))
		return
	}
	if j == nil {
		log.Printf("worker: job %s not found, skipping", jobID)
		return
	}

	tStart := time.Now()

	prInfo, err := w.prRepo.GetPR(ctx, item.Token, j.PR.Owner, j.PR.Repo, j.PR.Number)
	if err != nil {
		fail(fmt.Errorf("get PR: %w", err))
		return
	}

	t0 := time.Now()
	changed, err := w.prRepo.ListChangedGoFiles(ctx, item.Token, j.PR.Owner, j.PR.Repo, j.PR.Number)
	if err != nil {
		fail(fmt.Errorf("list changed files: %w", err))
		return
	}
	log.Printf("worker: ListChangedGoFiles took %s (%d files)", time.Since(t0), len(changed))

	t1 := time.Now()
	prepared, err := w.sourceTree.Prepare(ctx, *prInfo, item.Token, changed)
	if err != nil {
		fail(fmt.Errorf("prepare source: %w", err))
		return
	}
	defer func() { _ = prepared.Cleanup() }()
	log.Printf("worker: Prepare (clone+load) took %s", time.Since(t1))

	t2 := time.Now()
	graph, err := w.cgBuilder.Build(ctx, prepared.Packages, prepared.ChangedPackages)
	if err != nil {
		fail(fmt.Errorf("build callgraph: %w", err))
		return
	}
	log.Printf("worker: Build callgraph took %s (nodes=%d edges=%d)", time.Since(t2), len(graph.Nodes), len(graph.Edges))

	t3 := time.Now()
	result, err := w.clusterer.Cluster(ctx, graph)
	if err != nil {
		fail(fmt.Errorf("cluster: %w", err))
		return
	}
	log.Printf("worker: Cluster took %s", time.Since(t3))

	t4 := time.Now()
	diffFiles, err := analyzer.CollectDiffFiles(ctx, w.prRepo, item.Token, *prInfo, changed)
	if err != nil {
		log.Printf("worker: CollectDiffFiles error (non-fatal): %v", err)
		diffFiles = nil
	}
	log.Printf("worker: CollectDiffFiles took %s (%d files)", time.Since(t4), len(diffFiles))
	log.Printf("worker: total job %s took %s", jobID, time.Since(tStart))

	analysisID := domain.AnalysisID(w.newID())
	a := &domain.Analysis{
		ID:           analysisID,
		PR:           *prInfo,
		Result:       result,
		ChangedFiles: diffFiles,
		CreatedAt:    w.now().UTC(),
	}
	if err := w.analyses.Save(ctx, a); err != nil {
		fail(fmt.Errorf("save analysis: %w", err))
		return
	}

	if err := w.jobs.UpdateAnalysisID(ctx, jobID, analysisID); err != nil {
		fail(fmt.Errorf("update analysis_id: %w", err))
		return
	}
	if err := w.jobs.UpdateStatus(ctx, jobID, domain.JobStatusDone, ""); err != nil {
		log.Printf("worker: failed to mark job %s done: %v", jobID, err)
	}
}

func workerRandomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("workerRandomID: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
