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

// analysisResultJSON は analyses.result カラムに保存する JSON 構造。
// 旧形式（ClusterResult ベタ）との後方互換のため changed_files は omitempty にする。
type analysisResultJSON struct {
	ClusterResult domain.ClusterResult `json:"cluster_result"`
	ChangedFiles  []domain.DiffFile    `json:"changed_files,omitempty"`
}

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
	payload := analysisResultJSON{
		ClusterResult: *a.Result,
		ChangedFiles:  a.ChangedFiles,
	}
	resultJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal analysis result: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO analyses
		 (id, owner, repo, pr_number, pr_title, pr_base_ref, pr_head_ref, pr_base_sha, pr_head_sha, result, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(a.ID), a.PR.Owner, a.PR.Repo, a.PR.Number,
		a.PR.Title, a.PR.BaseRef, a.PR.HeadRef, a.PR.BaseSHA, a.PR.HeadSHA,
		string(resultJSON), a.CreatedAt.UTC(),
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
		prTitle    string
		prBaseRef  string
		prHeadRef  string
		prBaseSHA  string
		prHeadSHA  string
		resultJSON string
		createdAt  time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT owner, repo, pr_number, pr_title, pr_base_ref, pr_head_ref, pr_base_sha, pr_head_sha, result, created_at
		 FROM analyses WHERE id = ?`,
		string(id),
	).Scan(&owner, &repo, &prNumber, &prTitle, &prBaseRef, &prHeadRef, &prBaseSHA, &prHeadSHA, &resultJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query analysis: %w", err)
	}

	result, changedFiles, err := unmarshalAnalysisResult(resultJSON)
	if err != nil {
		return nil, fmt.Errorf("unmarshal analysis result: %w", err)
	}
	return &domain.Analysis{
		ID: id,
		PR: domain.PRInfo{
			Owner:   owner,
			Repo:    repo,
			Number:  prNumber,
			Title:   prTitle,
			BaseRef: prBaseRef,
			HeadRef: prHeadRef,
			BaseSHA: prBaseSHA,
			HeadSHA: prHeadSHA,
		},
		Result:       result,
		ChangedFiles: changedFiles,
		CreatedAt:    createdAt,
	}, nil
}

// FindByPR は同一 PR の最新解析結果を返す。存在しない場合は (nil, nil) を返す。
func (r *AnalysisRepo) FindByPR(ctx context.Context, pr domain.PRInfo) (*domain.Analysis, error) {
	var (
		id         string
		prTitle    string
		prBaseRef  string
		prHeadRef  string
		prBaseSHA  string
		prHeadSHA  string
		resultJSON string
		createdAt  time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT id, pr_title, pr_base_ref, pr_head_ref, pr_base_sha, pr_head_sha, result, created_at
		 FROM analyses WHERE owner=? AND repo=? AND pr_number=? ORDER BY created_at DESC LIMIT 1`,
		pr.Owner, pr.Repo, pr.Number,
	).Scan(&id, &prTitle, &prBaseRef, &prHeadRef, &prBaseSHA, &prHeadSHA, &resultJSON, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query analysis by PR: %w", err)
	}

	result, changedFiles, err := unmarshalAnalysisResult(resultJSON)
	if err != nil {
		return nil, fmt.Errorf("unmarshal analysis result: %w", err)
	}
	return &domain.Analysis{
		ID: domain.AnalysisID(id),
		PR: domain.PRInfo{
			Owner:   pr.Owner,
			Repo:    pr.Repo,
			Number:  pr.Number,
			Title:   prTitle,
			BaseRef: prBaseRef,
			HeadRef: prHeadRef,
			BaseSHA: prBaseSHA,
			HeadSHA: prHeadSHA,
		},
		Result:       result,
		ChangedFiles: changedFiles,
		CreatedAt:    createdAt,
	}, nil
}

// unmarshalAnalysisResult は新形式（{cluster_result, changed_files}）と
// 旧形式（ClusterResult ベタ）の両方を読めるようにする。
func unmarshalAnalysisResult(raw string) (*domain.ClusterResult, []domain.DiffFile, error) {
	// 新形式を試みる
	var payload analysisResultJSON
	if err := json.Unmarshal([]byte(raw), &payload); err == nil && payload.ClusterResult.Graph.Nodes != nil {
		return &payload.ClusterResult, payload.ChangedFiles, nil
	}
	// 旧形式（ClusterResult ベタ）にフォールバック
	var legacy domain.ClusterResult
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		return nil, nil, err
	}
	return &legacy, nil, nil
}
