package job

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/analyzer"
	"github.com/s-sh1m0/u-22_procon/backend/internal/infra/cluster"
)

// jobStore はワーカーが使うジョブ永続化インターフェース。
type jobStore interface {
	FindByID(ctx context.Context, id domain.JobID) (*domain.Job, error)
	UpdateStatus(ctx context.Context, id domain.JobID, status domain.JobStatus, errMsg string) error
	UpdatePhase(ctx context.Context, id domain.JobID, phase domain.JobPhase) error
	UpdateAnalysisID(ctx context.Context, id domain.JobID, analysisID domain.AnalysisID) error
}

// sourceTreePreparer はワーカーが使うソースツリー準備インターフェース。
type sourceTreePreparer interface {
	Prepare(ctx context.Context, pr domain.PRInfo, token string, changed []domain.ChangedFile) (*analyzer.PreparedSource, error)
	PrepareAtSHA(ctx context.Context, pr domain.PRInfo, token, sha string, changed []domain.ChangedFile) (*analyzer.PreparedSource, error)
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
	// setPhase はフロントの進捗表示用にジョブの現在フェーズを更新する。
	// 進捗表示は付随情報なので、失敗してもジョブ自体は止めずログのみ残す。
	setPhase := func(p domain.JobPhase) {
		if err := w.jobs.UpdatePhase(ctx, jobID, p); err != nil {
			log.Printf("worker: failed to set phase %s for job %s: %v", p, jobID, err)
		}
	}

	if err := w.jobs.UpdateStatus(ctx, jobID, domain.JobStatusRunning, ""); err != nil {
		log.Printf("worker: failed to mark job %s running: %v", jobID, err)
	}
	setPhase(domain.JobPhaseClone)

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

	// クローン完了後、最初にグラフ構築へ入った時点で一度だけ build_graph フェーズへ移す
	// (base/head が並列なため sync.Once で先着のみ反映する)。
	var buildOnce sync.Once
	onBuild := func() { buildOnce.Do(func() { setPhase(domain.JobPhaseBuildGraph) }) }

	t1 := time.Now()
	headGraph, baseGraph, cleanups, err := w.buildBothGraphs(ctx, *prInfo, item.Token, changed, onBuild)
	defer func() {
		for _, c := range cleanups {
			if c != nil {
				_ = c()
			}
		}
	}()
	if err != nil {
		if errors.Is(err, analyzer.ErrRepositoryTooLarge) {
			fail(errors.New("大規模リポジトリは未対応です（リポジトリ全体のパッケージ数が上限を超過）。より小規模な Go リポジトリの PR でお試しください。"))
			return
		}
		fail(fmt.Errorf("build base/head graphs: %w", err))
		return
	}
	log.Printf("worker: parallel base+head build took %s (head: nodes=%d edges=%d, base: nodes=%d edges=%d)",
		time.Since(t1),
		len(headGraph.Nodes), len(headGraph.Edges),
		len(baseGraph.Nodes), len(baseGraph.Edges))

	setPhase(domain.JobPhaseDiff)
	// base/head のグラフを合成して DiffStatus を付与する
	diffGraph := analyzer.MergeWithDiffStatus(baseGraph, headGraph)

	// 循環参照を検出（base にあった cycle と比較して IsNew を決定）
	cycles := analyzer.DetectNewCycles(baseGraph, headGraph)

	// 依存方向の逆転（domain → infra 等）を検出（added エッジを新規違反とする）
	violations := analyzer.DetectLayeringViolations(diffGraph)

	setPhase(domain.JobPhaseCluster)
	t3 := time.Now()
	result, err := w.clusterer.Cluster(ctx, diffGraph)
	if err != nil {
		fail(fmt.Errorf("cluster: %w", err))
		return
	}
	result.Cycles = cycles
	result.LayerViolations = violations
	log.Printf("worker: Cluster took %s (cycles=%d, violations=%d)", time.Since(t3), len(cycles), len(violations))

	setPhase(domain.JobPhaseVisualize)
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

// buildBothGraphs は head/base 両方の callgraph を errgroup で並列に構築する。
// 戻り値の cleanups スライスには各 PreparedSource.Cleanup が入る（呼び出し元が defer で全部呼ぶこと）。
// 一方が失敗した時点で他方の ctx もキャンセルされ、エラーが返る。
// onBuild は各 Prepare（クローン）完了後・Build 開始前に呼ばれる進捗通知コールバック。
func (w *Worker) buildBothGraphs(
	ctx context.Context,
	pr domain.PRInfo,
	token string,
	changed []domain.ChangedFile,
	onBuild func(),
) (head, base *domain.Graph, cleanups []func() error, err error) {
	cleanups = make([]func() error, 2)

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		prepared, err := w.sourceTree.Prepare(egCtx, pr, token, changed)
		if err != nil {
			return fmt.Errorf("head prepare: %w", err)
		}
		cleanups[0] = prepared.Cleanup
		onBuild()
		g, err := w.cgBuilder.Build(egCtx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedFileAbsPaths, prepared.RepoRoot)
		if err != nil {
			return fmt.Errorf("head build: %w", err)
		}
		head = g
		return nil
	})

	eg.Go(func() error {
		baseChanged := normalizeChangedForBase(changed)
		prepared, err := w.sourceTree.PrepareAtSHA(egCtx, pr, token, pr.BaseSHA, baseChanged)
		if err != nil {
			return fmt.Errorf("base prepare: %w", err)
		}
		cleanups[1] = prepared.Cleanup
		onBuild()
		g, err := w.cgBuilder.Build(egCtx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedFileAbsPaths, prepared.RepoRoot)
		if err != nil {
			return fmt.Errorf("base build: %w", err)
		}
		base = g
		return nil
	})

	if err = eg.Wait(); err != nil {
		return nil, nil, cleanups, err
	}
	return head, base, cleanups, nil
}

// normalizeChangedForBase は head 基準の changed リストを base 側のリポジトリで使える形に正規化する。
//   - Status=added は base 側に存在しないため除外
//   - Status=renamed は Filename を PreviousFilename に置き換える
//   - その他（modified, removed）はそのまま
func normalizeChangedForBase(changed []domain.ChangedFile) []domain.ChangedFile {
	out := make([]domain.ChangedFile, 0, len(changed))
	for _, f := range changed {
		switch f.Status {
		case domain.FileStatusAdded:
			continue
		case domain.FileStatusRenamed:
			cp := f
			if cp.PreviousFilename != "" {
				cp.Filename = cp.PreviousFilename
			}
			out = append(out, cp)
		default:
			out = append(out, f)
		}
	}
	return out
}

func workerRandomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("workerRandomID: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
