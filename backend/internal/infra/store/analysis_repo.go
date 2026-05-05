package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// AnalysisRepo は domain.AnalysisRepository の SQLite 実装
type AnalysisRepo struct {
	db *sql.DB
}

// NewAnalysisRepo は AnalysisRepo を返す。
func NewAnalysisRepo(db *sql.DB) *AnalysisRepo {
	return &AnalysisRepo{db: db}
}

// Save は解析結果を analyses テーブルに保存する。result は JSON でシリアライズする。
func (r *AnalysisRepo) Save(ctx context.Context, a *domain.Analysis) error {
	resultJSON, err := json.Marshal(a.Result)
	if err != nil {
		return fmt.Errorf("marshal analysis result: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO analyses (id, owner, repo, pr_number, result, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		string(a.ID), a.PR.Owner, a.PR.Repo, a.PR.Number, string(resultJSON), a.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert analysis: %w", err)
	}
	return nil
}

// FindByID は ID で解析結果を返す。存在しない場合は (nil, nil) を返す。
func (r *AnalysisRepo) FindByID(ctx context.Context, id domain.AnalysisID) (*domain.Analysis, error) {
	var (
		owner      string
		repo       string
		prNumber   int
		resultJSON string
		createdAt  time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT owner, repo, pr_number, result, created_at FROM analyses WHERE id = ?`,
		string(id),
	).Scan(&owner, &repo, &prNumber, &resultJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query analysis: %w", err)
	}

	var result domain.ClusterResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, fmt.Errorf("unmarshal analysis result: %w", err)
	}
	return &domain.Analysis{
		ID:        id,
		PR:        domain.PRInfo{Owner: owner, Repo: repo, Number: prNumber},
		Result:    &result,
		CreatedAt: createdAt,
	}, nil
}

// FindByPR は同一 PR の最新解析結果を返す。存在しない場合は (nil, nil) を返す。
func (r *AnalysisRepo) FindByPR(ctx context.Context, pr domain.PRInfo) (*domain.Analysis, error) {
	var (
		id         string
		resultJSON string
		createdAt  time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT id, result, created_at FROM analyses WHERE owner=? AND repo=? AND pr_number=? ORDER BY created_at DESC LIMIT 1`,
		pr.Owner, pr.Repo, pr.Number,
	).Scan(&id, &resultJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query analysis by PR: %w", err)
	}

	var result domain.ClusterResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		return nil, fmt.Errorf("unmarshal analysis result: %w", err)
	}
	return &domain.Analysis{
		ID:        domain.AnalysisID(id),
		PR:        pr,
		Result:    &result,
		CreatedAt: createdAt,
	}, nil
}
