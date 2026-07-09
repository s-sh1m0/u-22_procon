package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func TestAnalysisRepo_SaveAndFindByID(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	ctx := context.Background()

	result := &domain.ClusterResult{
		Clusters: []domain.Cluster{
			{ID: 0, Label: "pkg/foo", Nodes: []domain.NodeID{"fn:A", "fn:B"}},
		},
		Graph: domain.Graph{
			Nodes: []domain.Node{
				{ID: "fn:A", Name: "A", Package: "pkg/foo", File: "foo.go", Line: 10, Changed: true},
				{ID: "fn:B", Name: "B", Package: "pkg/foo", File: "foo.go", Line: 20, Changed: false},
			},
			Edges: []domain.Edge{{From: "fn:A", To: "fn:B"}},
		},
	}
	a := &domain.Analysis{
		ID: "test-analysis-id",
		PR: domain.PRInfo{
			Owner:   "owner",
			Repo:    "repo",
			Number:  42,
			Title:   "Add feature",
			BaseRef: "main",
			HeadRef: "feature-x",
			BaseSHA: "base123",
			HeadSHA: "head456",
		},
		Result:    result,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil analysis")
	}
	if got.PR.Owner != "owner" || got.PR.Repo != "repo" || got.PR.Number != 42 {
		t.Errorf("PR mismatch: got %+v", got.PR)
	}
	// PR メタ情報（タイトル・ref・SHA）が round-trip すること（GitHub 定義行リンクの ref に使う）。
	if got.PR.Title != "Add feature" || got.PR.BaseRef != "main" || got.PR.HeadRef != "feature-x" ||
		got.PR.BaseSHA != "base123" || got.PR.HeadSHA != "head456" {
		t.Errorf("PR metadata mismatch: got %+v", got.PR)
	}
	if len(got.Result.Clusters) != 1 || got.Result.Clusters[0].Label != "pkg/foo" {
		t.Errorf("Clusters mismatch: got %+v", got.Result.Clusters)
	}
	if len(got.Result.Graph.Nodes) != 2 {
		t.Errorf("Nodes count mismatch: got %d", len(got.Result.Graph.Nodes))
	}
	if len(got.Result.Graph.Edges) != 1 {
		t.Errorf("Edges count mismatch: got %d", len(got.Result.Graph.Edges))
	}
}

func TestAnalysisRepo_SaveAndFindByID_WithCycles(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	ctx := context.Background()

	result := &domain.ClusterResult{
		Graph: domain.Graph{
			Nodes: []domain.Node{
				{ID: "fn:A", DiffStatus: domain.DiffStatusAdded},
				{ID: "fn:B", DiffStatus: domain.DiffStatusRemoved},
			},
			Edges: []domain.Edge{{From: "fn:A", To: "fn:B", Status: domain.DiffStatusExisting}},
		},
		Cycles: []domain.Cycle{
			{ID: 0, Nodes: []domain.NodeID{"fn:A", "fn:B"}, IsNew: true},
		},
	}
	a := &domain.Analysis{
		ID:        "with-cycles",
		PR:        domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
		Result:    result,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.FindByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if len(got.Result.Cycles) != 1 || !got.Result.Cycles[0].IsNew {
		t.Errorf("cycles mismatch: %+v", got.Result.Cycles)
	}
	if got.Result.Graph.Nodes[0].DiffStatus != domain.DiffStatusAdded {
		t.Errorf("DiffStatus mismatch: %s", got.Result.Graph.Nodes[0].DiffStatus)
	}
	if got.Result.Graph.Edges[0].Status != domain.DiffStatusExisting {
		t.Errorf("Edge.Status mismatch: %s", got.Result.Graph.Edges[0].Status)
	}
}

func TestAnalysisRepo_FindByID_NotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	got, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestAnalysisRepo_FindByPR_ReturnsLatest(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	ctx := context.Background()
	pr := domain.PRInfo{Owner: "owner", Repo: "repo", Number: 1}

	emptyResult := &domain.ClusterResult{}

	old := &domain.Analysis{ID: "old", PR: pr, Result: emptyResult, CreatedAt: time.Now().UTC().Add(-time.Hour)}
	newerPR := pr
	newerPR.HeadSHA = "head789"
	newer := &domain.Analysis{ID: "newer", PR: newerPR, Result: emptyResult, CreatedAt: time.Now().UTC()}

	for _, a := range []*domain.Analysis{old, newer} {
		if err := repo.Save(ctx, a); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	got, err := repo.FindByPR(ctx, pr)
	if err != nil {
		t.Fatalf("FindByPR: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil analysis")
	}
	if got.ID != "newer" {
		t.Errorf("expected newest analysis, got id=%s", got.ID)
	}
	// 永続化された PR メタ情報（SHA）が引数 pr ではなく DB の値で復元されること。
	if got.PR.HeadSHA != "head789" {
		t.Errorf("expected persisted HeadSHA, got %q", got.PR.HeadSHA)
	}
}

// changed_files を別カラムに分離した現行形式で、Save→FindDiffByID が diff 本文を
// round-trip すること。かつ FindByID（グラフ経路）は diff 本文を読み込まないこと。
func TestAnalysisRepo_ChangedFilesRoundTrip(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	ctx := context.Background()

	a := &domain.Analysis{
		ID:     "a1",
		PR:     domain.PRInfo{Owner: "o", Repo: "r", Number: 1},
		Result: &domain.ClusterResult{Graph: domain.Graph{Nodes: []domain.Node{{ID: "fn:A"}}}},
		ChangedFiles: []domain.DiffFile{
			{Filename: "pkg/a.go", Status: domain.FileStatusAdded, Additions: 3, AfterContent: "package a"},
		},
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := repo.Save(ctx, a); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// diff 経路: changed_files が返る
	diff, err := repo.FindDiffByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("FindDiffByID: %v", err)
	}
	if diff == nil || len(diff.ChangedFiles) != 1 || diff.ChangedFiles[0].Filename != "pkg/a.go" ||
		diff.ChangedFiles[0].AfterContent != "package a" {
		t.Errorf("unexpected diff: %+v", diff)
	}

	// グラフ経路: Result は返るが diff 本文は読み込まない
	graph, err := repo.FindByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if graph == nil || len(graph.Result.Graph.Nodes) != 1 {
		t.Errorf("unexpected graph result: %+v", graph)
	}
	if graph.ChangedFiles != nil {
		t.Errorf("FindByID must not load diff bodies, got %+v", graph.ChangedFiles)
	}
}

func TestAnalysisRepo_FindDiffByID_NotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	got, err := repo.FindDiffByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

// PR-E 世代の行（result に {cluster_result, changed_files} を埋め込み・changed_files
// カラムは NULL）でも、FindDiffByID が result 埋め込みへフォールバックして diff を返し、
// FindByID / FindByPR がグラフを正しく復元すること。
func TestAnalysisRepo_LegacyEmbeddedFormat(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	embedded := analysisResultJSON{
		ClusterResult: domain.ClusterResult{Graph: domain.Graph{Nodes: []domain.Node{{ID: "fn:A"}}}},
		ChangedFiles:  []domain.DiffFile{{Filename: "old.go", Status: domain.FileStatusModified}},
	}
	raw, err := json.Marshal(embedded)
	if err != nil {
		t.Fatal(err)
	}
	// changed_files カラムを省いた（NULL の）レガシー行を直接挿入する。
	_, err = db.ExecContext(ctx,
		`INSERT INTO analyses (id, owner, repo, pr_number, result, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"legacy", "o", "r", 1, string(raw), time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	repo := NewAnalysisRepo(db)

	diff, err := repo.FindDiffByID(ctx, "legacy")
	if err != nil {
		t.Fatalf("FindDiffByID: %v", err)
	}
	if diff == nil || len(diff.ChangedFiles) != 1 || diff.ChangedFiles[0].Filename != "old.go" {
		t.Errorf("legacy diff fallback failed: %+v", diff)
	}

	graph, err := repo.FindByID(ctx, "legacy")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if graph == nil || len(graph.Result.Graph.Nodes) != 1 || graph.Result.Graph.Nodes[0].ID != "fn:A" {
		t.Errorf("legacy graph decode failed: %+v", graph)
	}

	cached, err := repo.FindByPR(ctx, domain.PRInfo{Owner: "o", Repo: "r", Number: 1})
	if err != nil {
		t.Fatalf("FindByPR: %v", err)
	}
	if cached == nil || len(cached.ChangedFiles) != 1 {
		t.Errorf("legacy FindByPR must recover changed files: %+v", cached)
	}
}

// 旧々形式（result が ClusterResult ベタ・diff 情報なし）でも graph が復元でき、
// diff は空で返ること。
func TestAnalysisRepo_LegacyBareFormat(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	raw, err := json.Marshal(domain.ClusterResult{Graph: domain.Graph{Nodes: []domain.Node{{ID: "fn:A"}}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.ExecContext(ctx,
		`INSERT INTO analyses (id, owner, repo, pr_number, result, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"bare", "o", "r", 1, string(raw), time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("insert bare row: %v", err)
	}

	repo := NewAnalysisRepo(db)

	graph, err := repo.FindByID(ctx, "bare")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if graph == nil || len(graph.Result.Graph.Nodes) != 1 {
		t.Errorf("bare graph decode failed: %+v", graph)
	}

	diff, err := repo.FindDiffByID(ctx, "bare")
	if err != nil {
		t.Fatalf("FindDiffByID: %v", err)
	}
	if diff == nil || diff.ChangedFiles != nil {
		t.Errorf("bare diff should be empty, got %+v", diff)
	}
}

func TestAnalysisRepo_FindByPR_NotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewAnalysisRepo(db)
	got, err := repo.FindByPR(context.Background(), domain.PRInfo{Owner: "o", Repo: "r", Number: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}
