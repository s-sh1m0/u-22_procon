package domain

import "time"

// JobID はジョブの一意識別子
type JobID string

// JobStatus はジョブの処理状態
type JobStatus string

const (
	JobStatusPending JobStatus = "pending"
	JobStatusRunning JobStatus = "running"
	JobStatusDone    JobStatus = "done"
	JobStatusError   JobStatus = "error"
)

// JobPhase は running 中のジョブの現在フェーズ。フロントの進捗ステップ表示に使う。
// pending のあいだは空文字、done/error 後は値を参照しない。
type JobPhase string

const (
	JobPhaseClone      JobPhase = "clone"       // リポジトリのクローン
	JobPhaseBuildGraph JobPhase = "build_graph" // 依存グラフの構築 (AST 解析)
	JobPhaseDiff       JobPhase = "diff"        // 差分との突き合わせ
	JobPhaseCluster    JobPhase = "cluster"     // クラスタ分類
	JobPhaseVisualize  JobPhase = "visualize"   // 可視化の生成
)

// Job は非同期解析ジョブのエンティティ
type Job struct {
	ID         JobID
	AnalysisID AnalysisID
	PR         PRInfo
	Status     JobStatus
	Phase      JobPhase
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
