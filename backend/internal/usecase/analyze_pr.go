package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// JobEnqueuer はジョブをキューに追加するインターフェース。infra/job.Queue が実装する。
type JobEnqueuer interface {
	Enqueue(jobID domain.JobID, token string)
}

// AnalyzePRUseCase はPR解析ジョブを開始するユースケース。
type AnalyzePRUseCase struct {
	analyses domain.AnalysisRepository
	jobs     domain.JobRepository
	queue    JobEnqueuer
	now      func() time.Time
	newID    func() string
}

// NewAnalyzePRUseCase は AnalyzePRUseCase を返す。
func NewAnalyzePRUseCase(analyses domain.AnalysisRepository, jobs domain.JobRepository, queue JobEnqueuer) *AnalyzePRUseCase {
	return &AnalyzePRUseCase{
		analyses: analyses,
		jobs:     jobs,
		queue:    queue,
		now:      time.Now,
		newID:    randomID,
	}
}

// Execute はPR解析ジョブを開始し、JobIDを返す。
// キャッシュ (同一 PR の既存解析) があれば done ジョブを即返し、enqueue しない。
func (uc *AnalyzePRUseCase) Execute(ctx context.Context, token string, pr domain.PRInfo) (domain.JobID, error) {
	cached, err := uc.analyses.FindByPR(ctx, pr)
	if err != nil {
		return "", fmt.Errorf("find cached analysis: %w", err)
	}

	now := uc.now().UTC()
	jobID := domain.JobID(uc.newID())

	if cached != nil && cached.ChangedFiles != nil {
		j := &domain.Job{
			ID:         jobID,
			AnalysisID: cached.ID,
			PR:         pr,
			Status:     domain.JobStatusDone,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := uc.jobs.Save(ctx, j); err != nil {
			return "", fmt.Errorf("save cached job: %w", err)
		}
		return jobID, nil
	}

	j := &domain.Job{
		ID:        jobID,
		PR:        pr,
		Status:    domain.JobStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := uc.jobs.Save(ctx, j); err != nil {
		return "", fmt.Errorf("save job: %w", err)
	}
	uc.queue.Enqueue(jobID, token)
	return jobID, nil
}

func randomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("randomID: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
