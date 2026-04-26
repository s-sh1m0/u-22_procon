package store

import (
	"context"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// AnalysisRepo は domain.AnalysisRepository の SQLite 実装
type AnalysisRepo struct {
	// TODO: *sql.DB を埋め込む
}

func (r *AnalysisRepo) Save(ctx context.Context, a *domain.Analysis) error {
	// TODO: implement
	return nil
}

func (r *AnalysisRepo) FindByID(ctx context.Context, id domain.AnalysisID) (*domain.Analysis, error) {
	// TODO: implement
	return nil, nil
}

func (r *AnalysisRepo) FindByPR(ctx context.Context, pr domain.PRInfo) (*domain.Analysis, error) {
	// TODO: implement
	return nil, nil
}
