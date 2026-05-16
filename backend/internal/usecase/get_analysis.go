package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// ErrJobNotFound はジョブが存在しない場合のエラー。
var ErrJobNotFound = errors.New("job not found")

// ErrJobNotReady はジョブがまだ完了していない場合のエラー。
var ErrJobNotReady = errors.New("job not ready")

// GetAnalysisUseCase は解析結果・ジョブ状態の取得ユースケース。
type GetAnalysisUseCase struct {
	analyses domain.AnalysisRepository
	jobs     domain.JobRepository
}

// NewGetAnalysisUseCase は GetAnalysisUseCase を返す。
func NewGetAnalysisUseCase(analyses domain.AnalysisRepository, jobs domain.JobRepository) *GetAnalysisUseCase {
	return &GetAnalysisUseCase{analyses: analyses, jobs: jobs}
}

// GetJob はジョブの状態を返す。存在しない場合は (nil, nil) を返す。
func (uc *GetAnalysisUseCase) GetJob(ctx context.Context, id domain.JobID) (*domain.Job, error) {
	j, err := uc.jobs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find job: %w", err)
	}
	return j, nil
}

// GetGraph はジョブに紐づく解析結果を返す。
// mode に応じてクラスタ分割を再構成する。ClusterModeLouvain（または空）は保存済みの Louvain 結果をそのまま返す。
// ジョブが存在しない場合は ErrJobNotFound、まだ完了していない場合は ErrJobNotReady を返す。
func (uc *GetAnalysisUseCase) GetGraph(ctx context.Context, id domain.JobID, mode domain.ClusterMode) (*domain.Analysis, error) {
	j, err := uc.jobs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find job: %w", err)
	}
	if j == nil {
		return nil, ErrJobNotFound
	}
	if j.Status != domain.JobStatusDone || j.AnalysisID == "" {
		return nil, ErrJobNotReady
	}

	a, err := uc.analyses.FindByID(ctx, j.AnalysisID)
	if err != nil {
		return nil, fmt.Errorf("find analysis: %w", err)
	}
	if a == nil {
		return nil, fmt.Errorf("data integrity error: job %s is done but analysis %s not found", id, j.AnalysisID)
	}
	// リポジトリが返す *Analysis を直接書き換えないようコピーしてから Result を差し替える。
	// 将来リポジトリがキャッシュ実装になっても呼び出し間で副作用が漏れないようにする防御策。
	out := *a
	out.Result = applyClusterMode(a.Result, mode)
	return &out, nil
}
