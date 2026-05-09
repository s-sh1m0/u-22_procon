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

// SessionRepository はログインセッションの永続化インターフェース
type SessionRepository interface {
	Save(ctx context.Context, s *Session) error
	FindByID(ctx context.Context, id SessionID) (*Session, error)
	Delete(ctx context.Context, id SessionID) error
}

// PRRepository は GitHub から PR メタ情報・変更ファイル一覧を取得するインターフェース。
// token はユーザーの OAuth アクセストークン（domain は oauth2 に依存しないため string で受ける）。
type PRRepository interface {
	GetPR(ctx context.Context, token, owner, repo string, number int) (*PRInfo, error)
	ListChangedGoFiles(ctx context.Context, token, owner, repo string, number int) ([]ChangedFile, error)
	// GetFileContent は指定 ref（branch/tag/SHA）のファイル全文を返す。
	// ファイルが存在しない場合は (nil, nil) を返す。
	GetFileContent(ctx context.Context, token, owner, repo, path, ref string) ([]byte, error)
}
