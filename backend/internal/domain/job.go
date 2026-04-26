package domain

import "time"

// JobID はジョブの一意識別子
type JobID string

// JobStatus はジョブの処理状態
type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusDone     JobStatus = "done"
	JobStatusError    JobStatus = "error"
)

// Job は非同期解析ジョブのエンティティ
type Job struct {
	ID         JobID
	AnalysisID AnalysisID
	PR         PRInfo
	Status     JobStatus
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
