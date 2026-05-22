package store

import (
	"context"
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
