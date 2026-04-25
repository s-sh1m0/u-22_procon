package store

import (
	"context"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// JobRepo は domain.JobRepository の SQLite 実装
type JobRepo struct {
	// TODO: *sql.DB を埋め込む
}

func (r *JobRepo) Save(ctx context.Context, j *domain.Job) error {
	// TODO: implement
	return nil
}

func (r *JobRepo) UpdateStatus(ctx context.Context, id domain.JobID, status domain.JobStatus, errMsg string) error {
	// TODO: implement
	return nil
}

func (r *JobRepo) FindByID(ctx context.Context, id domain.JobID) (*domain.Job, error) {
	// TODO: implement
	return nil, nil
}
