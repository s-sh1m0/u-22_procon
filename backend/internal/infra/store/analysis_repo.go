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

// analysisResultJSON は PR-E 世代の result カラム形式（ClusterResult と diff 本文を
// 単一 JSON にまとめた形）。現在の Save は ClusterResult と changed_files を別々に
// 保存するため書き込みでは使わず、旧形式行を読むための後方互換にのみ用いる。
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

// Save は解析結果を analyses テーブルに保存する。
// グラフ本体（ClusterResult）は result カラムに、diff 本文（ChangedFiles）は
// changed_files カラムに分離して保存する。これにより /api/graph の読み出しで
// diff 本文を unmarshal せずに済む。
func (r *AnalysisRepo) Save(ctx context.Context, a *domain.Analysis) error {
	resultJSON, err := json.Marshal(*a.Result)
	if err != nil {
		return fmt.Errorf("marshal cluster result: %w", err)
	}
	// ChangedFiles が nil のとき（diff 収集に失敗した等）は NULL を保存し、
	// キャッシュ判定（ChangedFiles != nil）で「diff 未取得」として扱えるようにする。
	var changedCol any
	if a.ChangedFiles != nil {
		changedJSON, err := json.Marshal(a.ChangedFiles)
		if err != nil {
			return fmt.Errorf("marshal changed files: %w", err)
		}
		changedCol = string(changedJSON)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO analyses
		 (id, owner, repo, pr_number, pr_title, pr_base_ref, pr_head_ref, pr_base_sha, pr_head_sha, result, changed_files, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		string(a.ID), a.PR.Owner, a.PR.Repo, a.PR.Number,
		a.PR.Title, a.PR.BaseRef, a.PR.HeadRef, a.PR.BaseSHA, a.PR.HeadSHA,
		string(resultJSON), changedCol, a.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert analysis: %w", err)
	}
	return nil
}

// FindByID は ID でグラフ結果を返す。diff 本文（ChangedFiles）は読み込まない
// （/api/graph の軽量読み出し経路）。存在しない場合は (nil, nil) を返す。
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

	result, err := unmarshalClusterResult(resultJSON)
	if err != nil {
		return nil, fmt.Errorf("unmarshal cluster result: %w", err)
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
		Result:    result,
		CreatedAt: createdAt,
	}, nil
}

// FindDiffByID は ID で diff 本文（ChangedFiles）のみを返す（/api/diff の読み出し経路）。
// グラフ本体（Result）は読み込まない。存在しない場合は (nil, nil) を返す。
func (r *AnalysisRepo) FindDiffByID(ctx context.Context, id domain.AnalysisID) (*domain.Analysis, error) {
	var (
		resultJSON string
		changedCol sql.NullString
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT result, changed_files FROM analyses WHERE id = ?`,
		string(id),
	).Scan(&resultJSON, &changedCol)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query analysis diff: %w", err)
	}

	changedFiles, err := unmarshalChangedFiles(resultJSON, changedCol)
	if err != nil {
		return nil, fmt.Errorf("unmarshal changed files: %w", err)
	}
	return &domain.Analysis{ID: id, ChangedFiles: changedFiles}, nil
}

// FindByPR は同一 PR の最新解析結果を返す。キャッシュ判定に使うため
// diff 本文（ChangedFiles）も復元する。存在しない場合は (nil, nil) を返す。
func (r *AnalysisRepo) FindByPR(ctx context.Context, pr domain.PRInfo) (*domain.Analysis, error) {
	var (
		id         string
		prTitle    string
		prBaseRef  string
		prHeadRef  string
		prBaseSHA  string
		prHeadSHA  string
		resultJSON string
		changedCol sql.NullString
		createdAt  time.Time
	)
	err := r.db.QueryRowContext(ctx,
		`SELECT id, pr_title, pr_base_ref, pr_head_ref, pr_base_sha, pr_head_sha, result, changed_files, created_at
		 FROM analyses WHERE owner=? AND repo=? AND pr_number=? ORDER BY created_at DESC LIMIT 1`,
		pr.Owner, pr.Repo, pr.Number,
	).Scan(&id, &prTitle, &prBaseRef, &prHeadRef, &prBaseSHA, &prHeadSHA, &resultJSON, &changedCol, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query analysis by PR: %w", err)
	}

	result, err := unmarshalClusterResult(resultJSON)
	if err != nil {
		return nil, fmt.Errorf("unmarshal cluster result: %w", err)
	}
	changedFiles, err := unmarshalChangedFiles(resultJSON, changedCol)
	if err != nil {
		return nil, fmt.Errorf("unmarshal changed files: %w", err)
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

// unmarshalClusterResult は result カラムから ClusterResult を取り出す。
// diff 本文は無視するため、グラフ読み経路で diff を unmarshal しない。
//   - 現行形式: result は ClusterResult ベタ（diff は changed_files カラム）
//   - PR-E 形式: result は {cluster_result, changed_files} なので cluster_result を取り出す
//   - 旧々形式: result は ClusterResult ベタ
func unmarshalClusterResult(raw string) (*domain.ClusterResult, error) {
	// PR-E のラッパー形式を試す。
	var wrapper analysisResultJSON
	if err := json.Unmarshal([]byte(raw), &wrapper); err == nil && wrapper.ClusterResult.Graph.Nodes != nil {
		return &wrapper.ClusterResult, nil
	}
	// ベタの ClusterResult にフォールバック（現行形式 / 旧々形式）。
	var cr domain.ClusterResult
	if err := json.Unmarshal([]byte(raw), &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// unmarshalChangedFiles は diff 本文を取り出す。
// changed_files カラムがあればそこから、無ければ PR-E 形式の result 埋め込みから復元する。
func unmarshalChangedFiles(resultRaw string, changedCol sql.NullString) ([]domain.DiffFile, error) {
	if changedCol.Valid {
		var files []domain.DiffFile
		if err := json.Unmarshal([]byte(changedCol.String), &files); err != nil {
			return nil, err
		}
		return files, nil
	}
	// PR-E 形式: result に changed_files が埋め込まれている。
	var wrapper analysisResultJSON
	if err := json.Unmarshal([]byte(resultRaw), &wrapper); err == nil && wrapper.ClusterResult.Graph.Nodes != nil {
		return wrapper.ChangedFiles, nil
	}
	// 旧々形式: diff データ無し。
	return nil, nil
}
