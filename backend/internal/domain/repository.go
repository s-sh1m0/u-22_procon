package domain

import "context"

// AnalysisRepository は解析結果の永続化インターフェース
type AnalysisRepository interface {
	Save(ctx context.Context, a *Analysis) error
	FindByID(ctx context.Context, id AnalysisID) (*Analysis, error)
	FindByPR(ctx context.Context, pr PRInfo) (*Analysis, error)
}

// JobRepository はジョブ状態の永続化インターフェース
type JobRepository interface {
	Save(ctx context.Context, j *Job) error
	UpdateStatus(ctx context.Context, id JobID, status JobStatus, errMsg string) error
	FindByID(ctx context.Context, id JobID) (*Job, error)
}
