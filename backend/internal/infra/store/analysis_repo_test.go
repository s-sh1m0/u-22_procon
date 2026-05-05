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
	defer db.Close()

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
		ID:        "test-analysis-id",
		PR:        domain.PRInfo{Owner: "owner", Repo: "repo", Number: 42},
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

func TestAnalysisRepo_FindByID_NotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

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
	defer db.Close()

	repo := NewAnalysisRepo(db)
	ctx := context.Background()
	pr := domain.PRInfo{Owner: "owner", Repo: "repo", Number: 1}

	emptyResult := &domain.ClusterResult{}

	old := &domain.Analysis{ID: "old", PR: pr, Result: emptyResult, CreatedAt: time.Now().UTC().Add(-time.Hour)}
	newer := &domain.Analysis{ID: "newer", PR: pr, Result: emptyResult, CreatedAt: time.Now().UTC()}

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
}

func TestAnalysisRepo_FindByPR_NotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewAnalysisRepo(db)
	got, err := repo.FindByPR(context.Background(), domain.PRInfo{Owner: "o", Repo: "r", Number: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}
