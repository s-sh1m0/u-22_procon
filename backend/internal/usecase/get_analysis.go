package usecase

import (
	"context"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// GetAnalysisUseCase は解析結果・ジョブ状態の取得ユースケース
type GetAnalysisUseCase struct {
	// TODO: analysisRepo, jobRepo を注入
}

// GetJob はジョブの状態を返す
func (uc *GetAnalysisUseCase) GetJob(ctx context.Context, id domain.JobID) (*domain.Job, error) {
	// TODO: implement
	return nil, nil
}

// GetAnalysis は解析結果を返す
func (uc *GetAnalysisUseCase) GetAnalysis(ctx context.Context, id domain.AnalysisID) (*domain.Analysis, error) {
	// TODO: implement
	return nil, nil
}
