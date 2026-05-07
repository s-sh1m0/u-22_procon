package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// JobRepo は domain.JobRepository の SQLite 実装
type JobRepo struct {
	db *sql.DB
}

// NewJobRepo は JobRepo を返す。
func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

// Save はジョブを jobs テーブルに保存する。AnalysisID が空のときは NULL を挿入する。
func (r *JobRepo) Save(ctx context.Context, j *domain.Job) error {
	var analysisID sql.NullString
	if j.AnalysisID != "" {
		analysisID = sql.NullString{String: string(j.AnalysisID), Valid: true}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO jobs (id, analysis_id, status, error, created_at, updated_at,
		 pr_owner, pr_repo, pr_number, pr_title, pr_base_ref, pr_head_ref, pr_head_sha)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(j.ID), analysisID, string(j.Status), j.Error,
		j.CreatedAt.UTC(), j.UpdatedAt.UTC(),
		j.PR.Owner, j.PR.Repo, j.PR.Number,
		j.PR.Title, j.PR.BaseRef, j.PR.HeadRef, j.PR.HeadSHA,
	)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

// UpdateStatus はジョブのステータスとエラーメッセージを更新する。
func (r *JobRepo) UpdateStatus(ctx context.Context, id domain.JobID, status domain.JobStatus, errMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE jobs SET status=?, error=?, updated_at=? WHERE id=?`,
		string(status), errMsg, time.Now().UTC(), string(id),
	)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}
	return nil
}

// UpdateAnalysisID はジョブの analysis_id を更新する。ワーカーが解析完了時に呼ぶ。
// このメソッドは domain.JobRepository インターフェースには含まれない（ワーカー専用）。
func (r *JobRepo) UpdateAnalysisID(ctx context.Context, id domain.JobID, analysisID domain.AnalysisID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE jobs SET analysis_id=?, updated_at=? WHERE id=?`,
		string(analysisID), time.Now().UTC(), string(id),
	)
	if err != nil {
		return fmt.Errorf("update job analysis_id: %w", err)
	}
	return nil
}

// MarkStaleJobsError はサーバ起動時に pending/running のままになっているジョブを
// error 状態にする。再起動でインメモリキューが消えて処理されないジョブを残さないため。
func (r *JobRepo) MarkStaleJobsError(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE jobs SET status='error', error='server restarted', updated_at=?
		 WHERE status IN ('pending', 'running')`,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("mark stale jobs: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// FindByID は ID でジョブを返す。存在しない場合は (nil, nil) を返す。
func (r *JobRepo) FindByID(ctx context.Context, id domain.JobID) (*domain.Job, error) {
	var (
		analysisID sql.NullString
		status     string
		errMsg     sql.NullString
		createdAt  time.Time
		updatedAt  time.Time
		prOwner    string
		prRepo     string
		prNumber   int
		prTitle    string
		prBaseRef  string
		prHeadRef  string
		prHeadSHA  string
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT analysis_id, status, error, created_at, updated_at,
		 pr_owner, pr_repo, pr_number, pr_title, pr_base_ref, pr_head_ref, pr_head_sha
		 FROM jobs WHERE id = ?`,
		string(id),
	).Scan(
		&analysisID, &status, &errMsg, &createdAt, &updatedAt,
		&prOwner, &prRepo, &prNumber, &prTitle, &prBaseRef, &prHeadRef, &prHeadSHA,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query job: %w", err)
	}

	j := &domain.Job{
		ID:        id,
		Status:    domain.JobStatus(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		PR: domain.PRInfo{
			Owner:   prOwner,
			Repo:    prRepo,
			Number:  prNumber,
			Title:   prTitle,
			BaseRef: prBaseRef,
			HeadRef: prHeadRef,
			HeadSHA: prHeadSHA,
		},
	}
	if analysisID.Valid {
		j.AnalysisID = domain.AnalysisID(analysisID.String)
	}
	if errMsg.Valid {
		j.Error = errMsg.String
	}
	return j, nil
}
