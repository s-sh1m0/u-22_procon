package job

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
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
// CloneBoth で head/base を1リポジトリ共有でcloneし、各worktreeを PrepareFromDir で
// ロードする（cloneとload/buildを分離し、cloneのみ共有・load/buildは並列を維持する）。
type sourceTreePreparer interface {
	CloneBoth(ctx context.Context, pr domain.PRInfo, token string) (*analyzer.ClonedWorktrees, error)
	PrepareFromDir(ctx context.Context, rootDir string, changed []domain.ChangedFile) (*analyzer.PreparedSource, error)
}

// defaultJobTimeout は1ジョブの解析にかける時間の上限のデフォルト値。
// 巨大リポジトリ（k8s 等）の解析が暴走してワーカーを無限に占有し、
// 後続ジョブ（＝他ユーザー）を巻き添えにするのを防ぐ安全弁。
const defaultJobTimeout = 120 * time.Second

// Worker はキューからジョブを受け取り解析パイプラインを実行する。
// Run は複数 goroutine から同時に呼び出して構わない（フィールドは不変、
// ジョブ処理は注入された依存のみを使い受信側の可変状態を持たない）。
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
	jobTimeout time.Duration
}

// Option は Worker の任意設定を変更する関数オプション。
type Option func(*Worker)

// WithJobTimeout は1ジョブあたりの処理時間上限を設定する。
// d <= 0 のときはデフォルト値を使う。
func WithJobTimeout(d time.Duration) Option {
	return func(w *Worker) {
		if d > 0 {
			w.jobTimeout = d
		}
	}
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
	opts ...Option,
) *Worker {
	w := &Worker{
		queue:      queue,
		jobs:       jobs,
		analyses:   analyses,
		prRepo:     prRepo,
		sourceTree: sourceTree,
		cgBuilder:  cgBuilder,
		clusterer:  clusterer,
		now:        time.Now,
		newID:      workerRandomID,
		jobTimeout: defaultJobTimeout,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
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

	// ジョブ単位のタイムアウト。暴走解析が1本のワーカーを占有し続けるのを防ぐ。
	timeout := w.jobTimeout
	if timeout <= 0 {
		timeout = defaultJobTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	fail := func(err error) {
		// タイムアウト由来の失敗は原因が分かる文言に置き換える（巨大リポジトリの可能性）。
		if errors.Is(err, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("解析がタイムアウトしました（%s）。対象リポジトリ／PR が大きすぎる可能性があります", timeout)
		}
		log.Printf("worker: job %s failed: %v", jobID, err)
		// 親 ctx のキャンセル／タイムアウトとは切り離しつつ、DB が詰まっても
		// シャットダウン時にワーカーが永久ブロックしないよう短いタイムアウトを付ける。
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = w.jobs.UpdateStatus(cleanupCtx, jobID, domain.JobStatusError, err.Error())
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

	// PR のメタ情報（Title / refs / SHA）は usecase.Execute が GetPR で解決済みで
	// jobs テーブルに永続化されているため、ここで再取得はせず j.PR をそのまま使う。
	t0 := time.Now()
	changed, err := w.prRepo.ListChangedGoFiles(ctx, item.Token, j.PR.Owner, j.PR.Repo, j.PR.Number)
	if err != nil {
		fail(fmt.Errorf("list changed files: %w", err))
		return
	}
	log.Printf("worker: ListChangedGoFiles took %s (%d files)", time.Since(t0), len(changed))

	// clone（共有）完了後、グラフ構築フェーズへ移す。
	onCloneDone := func() { setPhase(domain.JobPhaseBuildGraph) }

	t1 := time.Now()
	headGraph, baseGraph, cleanup, err := w.buildBothGraphs(ctx, j.PR, item.Token, changed, onCloneDone)
	defer func() {
		if cleanup != nil {
			_ = cleanup()
		}
	}()
	if err != nil {
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
	diffFiles, err := analyzer.CollectDiffFiles(ctx, w.prRepo, item.Token, j.PR, changed)
	if err != nil {
		log.Printf("worker: CollectDiffFiles error (non-fatal): %v", err)
		diffFiles = nil
	}
	log.Printf("worker: CollectDiffFiles took %s (%d files)", time.Since(t4), len(diffFiles))
	log.Printf("worker: total job %s took %s", jobID, time.Since(tStart))

	analysisID := domain.AnalysisID(w.newID())
	a := &domain.Analysis{
		ID:           analysisID,
		PR:           j.PR,
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

// buildBothGraphs は head/base 両方の callgraph を構築する。
// clone は CloneBoth で1リポジトリ共有（1回の fetch）にまとめ、その後の
// load（PrepareFromDir）と Build は errgroup で base/head 並列に実行する。
// 戻り値の cleanup は共有tmp dirを削除する単一のクロージャ（呼び出し元が defer で呼ぶこと）。
// 一方が失敗した時点で他方の ctx もキャンセルされ、エラーが返る。
// onCloneDone は clone 完了後・Build 開始前に一度だけ呼ばれる進捗通知コールバック。
func (w *Worker) buildBothGraphs(
	ctx context.Context,
	pr domain.PRInfo,
	token string,
	changed []domain.ChangedFile,
	onCloneDone func(),
) (head, base *domain.Graph, cleanup func() error, err error) {
	// base/head を1リポジトリ共有で clone（fetch/オブジェクトストアを共有）。
	cw, err := w.sourceTree.CloneBoth(ctx, pr, token)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("clone base/head: %w", err)
	}
	cleanup = cw.Cleanup
	onCloneDone()

	headDir, ok := cw.Dirs[pr.HeadSHA]
	if !ok {
		return nil, nil, cleanup, fmt.Errorf("clone base/head: missing worktree for head %s", pr.HeadSHA)
	}
	baseDir, ok := cw.Dirs[pr.BaseSHA]
	if !ok {
		return nil, nil, cleanup, fmt.Errorf("clone base/head: missing worktree for base %s", pr.BaseSHA)
	}

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		prepared, err := w.sourceTree.PrepareFromDir(egCtx, headDir, changed)
		if err != nil {
			return fmt.Errorf("head prepare: %w", err)
		}
		g, err := w.cgBuilder.Build(egCtx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedFileAbsPaths, prepared.RepoRoot)
		if err != nil {
			return fmt.Errorf("head build: %w", err)
		}
		head = g
		return nil
	})

	eg.Go(func() error {
		baseChanged := domain.NormalizeChangedForBase(changed)
		prepared, err := w.sourceTree.PrepareFromDir(egCtx, baseDir, baseChanged)
		if err != nil {
			return fmt.Errorf("base prepare: %w", err)
		}
		g, err := w.cgBuilder.Build(egCtx, prepared.Packages, prepared.ChangedPackages, prepared.ChangedFileAbsPaths, prepared.RepoRoot)
		if err != nil {
			return fmt.Errorf("base build: %w", err)
		}
		base = g
		return nil
	})

	if err = eg.Wait(); err != nil {
		return nil, nil, cleanup, err
	}
	return head, base, cleanup, nil
}

func workerRandomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("workerRandomID: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
