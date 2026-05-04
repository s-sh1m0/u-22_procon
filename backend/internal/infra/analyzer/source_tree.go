package analyzer

import (
	"context"
	"fmt"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// PreparedSource はclone・ロード・変更パッケージ特定の結果をまとめたもの。
type PreparedSource struct {
	RootDir         string
	Packages        []*packages.Package
	LoadErrors      []packages.Error
	ChangedPackages []string // パッケージID（重複なし・ソート済み）
	Cleanup         func() error
}

// SourceTree はCloneとPackageLoadを組み合わせたファサード。
// job worker (#5/#7) から最終的に呼ばれる入口を想定している。
type SourceTree struct {
	Cloner Cloner
	Loader PackageLoader
}

// NewSourceTree はSourceTreeを初期化する。
func NewSourceTree(c Cloner, l PackageLoader) *SourceTree {
	return &SourceTree{Cloner: c, Loader: l}
}

// Prepare は clone → load → 変更パッケージ特定 を一気に実施する。
// Loadが失敗したときはCloneのCleanupを内部で呼んでから返す（呼び出し側に部分状態を渡さない）。
// 成功時のみPreparedSource.Cleanupが返り、呼び出し側がdeferで呼ぶ責務を持つ。
func (s *SourceTree) Prepare(ctx context.Context, pr domain.PRInfo, token string, changed []domain.ChangedFile) (*PreparedSource, error) {
	cloned, err := s.Cloner.Clone(ctx, CloneRequest{
		Owner: pr.Owner,
		Repo:  pr.Repo,
		SHA:   pr.HeadSHA,
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("analyzer: clone %s/%s@%s: %w", pr.Owner, pr.Repo, pr.HeadSHA, err)
	}

	result, err := s.Loader.Load(ctx, cloned.RootDir)
	if err != nil {
		_ = cloned.Cleanup()
		return nil, fmt.Errorf("analyzer: load packages at %s: %w", cloned.RootDir, err)
	}

	changedPkgs := IdentifyChangedPackages(cloned.RootDir, result.Packages, changed)

	return &PreparedSource{
		RootDir:         cloned.RootDir,
		Packages:        result.Packages,
		LoadErrors:      result.Errors,
		ChangedPackages: changedPkgs,
		Cleanup:         cloned.Cleanup,
	}, nil
}
