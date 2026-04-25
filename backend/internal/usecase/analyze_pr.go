package usecase

import (
	"context"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// AnalyzePRUseCase はPR解析ユースケースの依存インターフェース
type AnalyzePRUseCase struct {
	// TODO: analyzer, cluster, github, analysisRepo, jobRepo を注入
}

// Execute はPR解析ジョブを開始し、JobIDを返す
func (uc *AnalyzePRUseCase) Execute(ctx context.Context, pr domain.PRInfo) (domain.JobID, error) {
	// TODO: implement
	// 1. キャッシュ確認（同一PRの既存結果があれば即返す）
	// 2. Jobエンティティ作成・保存
	// 3. ジョブキューに enqueue
	return "", nil
}
