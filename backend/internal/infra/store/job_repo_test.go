package store

import (
	"context"
	"testing"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func makeJob(id string, status domain.JobStatus) *domain.Job {
	now := time.Now().UTC().Truncate(time.Second)
	return &domain.Job{
		ID:     domain.JobID(id),
		Status: status,
		PR: domain.PRInfo{
			Owner:   "owner",
			Repo:    "repo",
			Number:  42,
			Title:   "fix: bug",
			BaseRef: "main",
			HeadRef: "fix-branch",
			HeadSHA: "abc123",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestJobRepo_SaveAndFindByID(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewJobRepo(db)
	ctx := context.Background()

	j := makeJob("job-1", domain.JobStatusPending)
	if err := repo.Save(ctx, j); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, j.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil job")
	}
	if got.ID != j.ID {
		t.Errorf("ID mismatch: got %s", got.ID)
	}
	if got.Status != domain.JobStatusPending {
		t.Errorf("Status mismatch: got %s", got.Status)
	}
	if got.AnalysisID != "" {
		t.Errorf("expected empty AnalysisID, got %s", got.AnalysisID)
	}
	if got.PR.Owner != "owner" || got.PR.Repo != "repo" || got.PR.Number != 42 {
		t.Errorf("PR mismatch: got %+v", got.PR)
	}
	if got.PR.Title != "fix: bug" || got.PR.HeadSHA != "abc123" {
		t.Errorf("PR detail mismatch: got %+v", got.PR)
	}
}

func TestJobRepo_FindByID_NotFound(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewJobRepo(db)
	got, err := repo.FindByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestJobRepo_UpdateStatus(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewJobRepo(db)
	ctx := context.Background()

	j := makeJob("job-2", domain.JobStatusPending)
	if err := repo.Save(ctx, j); err != nil {
		t.Fatal(err)
	}

	prevUpdatedAt := j.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	if err := repo.UpdateStatus(ctx, j.ID, domain.JobStatusRunning, ""); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.FindByID(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.JobStatusRunning {
		t.Errorf("expected running, got %s", got.Status)
	}
	if !got.UpdatedAt.After(prevUpdatedAt) {
		t.Errorf("updated_at did not advance")
	}
}

func TestJobRepo_UpdateStatus_WithError(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewJobRepo(db)
	ctx := context.Background()

	j := makeJob("job-3", domain.JobStatusRunning)
	if err := repo.Save(ctx, j); err != nil {
		t.Fatal(err)
	}

	if err := repo.UpdateStatus(ctx, j.ID, domain.JobStatusError, "something went wrong"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.FindByID(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.JobStatusError {
		t.Errorf("expected error status, got %s", got.Status)
	}
	if got.Error != "something went wrong" {
		t.Errorf("expected error message, got %q", got.Error)
	}
}

func TestJobRepo_UpdateAnalysisID(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewJobRepo(db)
	ctx := context.Background()

	j := makeJob("job-4", domain.JobStatusRunning)
	if err := repo.Save(ctx, j); err != nil {
		t.Fatal(err)
	}

	analysisID := domain.AnalysisID("analysis-xyz")
	if err := repo.UpdateAnalysisID(ctx, j.ID, analysisID); err != nil {
		t.Fatalf("UpdateAnalysisID: %v", err)
	}

	got, err := repo.FindByID(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.AnalysisID != analysisID {
		t.Errorf("expected analysisID=%s, got %s", analysisID, got.AnalysisID)
	}
}

func TestJobRepo_SaveWithAnalysisID(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	repo := NewJobRepo(db)
	ctx := context.Background()

	j := makeJob("job-5", domain.JobStatusDone)
	j.AnalysisID = "analysis-abc"

	if err := repo.Save(ctx, j); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, j.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.AnalysisID != "analysis-abc" {
		t.Errorf("expected analysis-abc, got %s", got.AnalysisID)
	}
}
