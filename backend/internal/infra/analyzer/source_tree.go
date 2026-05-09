package analyzer

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// PreparedSource はclone・ロード・変更パッケージ特定の結果をまとめたもの。
type PreparedSource struct {
	RootDir             string
	Packages            []*packages.Package
	LoadErrors          []packages.Error
	ChangedPackages     []string // パッケージID（重複なし・ソート済み）
	ChangedFileAbsPaths []string // PR で変更された .go ファイルの絶対パス（HEAD 側）
	Cleanup             func() error
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

// Prepare は clone → fast load → 変更パッケージ特定 → 近傍フルロード を一気に実施する。
//
// 2フェーズロードにより、プロジェクト内の変更パッケージとその近傍のみを
// 型情報付きでロードし、外部ライブラリを含む全パッケージのロードを回避する。
//
// Loadが失敗したときはCloneのCleanupを内部で呼んでから返す（呼び出し側に部分状態を渡さない）。
// 成功時のみPreparedSource.Cleanupが返り、呼び出し側がdeferで呼ぶ責務を持つ。
// モノレポ対応: cloneルートでパッケージが見つからない場合、go.mod を持つサブディレクトリを自動検出する。
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

	// go.mod を持つサブディレクトリを特定し、そこを基点にパッケージをロードする。
	// clonedRoot はリポジトリルート（GitHub API の相対パス基点）、loadDir は go.mod の親。
	clonedRoot := cloned.RootDir
	loadDir, err := findGoModRoot(clonedRoot)
	if err != nil {
		_ = cloned.Cleanup()
		return nil, fmt.Errorf("analyzer: find go.mod in %s: %w", clonedRoot, err)
	}

	// Phase 1: 型チェックなしで全プロジェクトパッケージの構造を高速ロード。
	// import グラフのみ取得し、変更パッケージとその近傍を特定する。
	t0 := time.Now()
	fastPkgs, err := s.Loader.FastLoad(ctx, loadDir)
	if err != nil {
		_ = cloned.Cleanup()
		return nil, fmt.Errorf("analyzer: fast-load packages at %s: %w", loadDir, err)
	}
	log.Printf("analyzer: phase1 FastLoad(%d pkgs) took %s", len(fastPkgs), time.Since(t0))

	// GitHub API の changed ファイルパスはリポジトリルート相対なので clonedRoot を基点にする。
	changedPkgs := IdentifyChangedPackages(clonedRoot, fastPkgs, changed)
	log.Printf("analyzer: changed packages: %v", changedPkgs)

	// HEAD に存在するファイルの絶対パスを収集する（removed は HEAD に存在しないため除外）。
	changedFileAbsPaths := make([]string, 0, len(changed))
	for _, f := range changed {
		if f.Status == "removed" {
			continue
		}
		changedFileAbsPaths = append(changedFileAbsPaths, filepath.Clean(filepath.Join(clonedRoot, f.Filename)))
	}

	if len(changedPkgs) == 0 {
		// 変更されたGoパッケージがない（非Goファイルのみの変更など）
		return &PreparedSource{
			RootDir:             loadDir,
			ChangedPackages:     nil,
			ChangedFileAbsPaths: changedFileAbsPaths,
			Cleanup:             cloned.Cleanup,
		}, nil
	}

	// Phase 2: 変更パッケージとその近傍（defaultMaxDepth ホップ以内）のみを
	// 型情報付きでロードする。外部ライブラリは型チェックの依存として読まれるが
	// 返却パッケージには含まれない。
	neighborhood := FindNeighborhood(changedPkgs, fastPkgs, defaultMaxDepth)
	log.Printf("analyzer: neighborhood(%d pkgs): %v", len(neighborhood), neighborhood)

	t1 := time.Now()
	result, err := s.Loader.Load(ctx, loadDir, neighborhood)
	if err != nil {
		_ = cloned.Cleanup()
		return nil, fmt.Errorf("analyzer: load packages at %s: %w", loadDir, err)
	}
	log.Printf("analyzer: phase2 Load(%d pkgs) took %s", len(result.Packages), time.Since(t1))

	return &PreparedSource{
		RootDir:             loadDir,
		Packages:            result.Packages,
		LoadErrors:          result.Errors,
		ChangedPackages:     changedPkgs,
		ChangedFileAbsPaths: changedFileAbsPaths,
		Cleanup:             cloned.Cleanup,
	}, nil
}

// findGoModRoot はルートから go.mod を探して最も浅い（ルートに近い）ディレクトリを返す。
// ルート直下に go.mod があればそのまま返す。
func findGoModRoot(root string) (string, error) {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		return root, nil
	}

	var found []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			found = append(found, filepath.Dir(path))
		}
		return nil
	})

	if len(found) == 0 {
		return "", fmt.Errorf("%w: no go.mod found under %s", ErrNoPackages, root)
	}
	// 最も浅い（パス区切りが最小の）ディレクトリを優先
	shallowest := found[0]
	for _, d := range found[1:] {
		if strings.Count(d, string(filepath.Separator)) < strings.Count(shallowest, string(filepath.Separator)) {
			shallowest = d
		}
	}
	return shallowest, nil
}
