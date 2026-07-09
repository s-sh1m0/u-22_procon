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

// defaultMaxNeighborhood は Phase 2 で型情報付きロードする近傍パッケージ数の上限。
// これを超える近傍は近いホップ優先で打ち切る。k8s のようなハブパッケージを含む
// 巨大リポジトリでフルロードがメモリ枯渇するのを防ぐためのガード。
// terraform（178 pkgs）で170近傍→タイムアウトした実測値から 150 に設定。
const defaultMaxNeighborhood = 150

// defaultNeighborhoodDepth はパッケージ近傍を辿る import グラフ上のホップ数上限。
// callgraph BFS の defaultMaxDepth（関数レベル）とは独立。ハブパッケージを含む
// リポジトリでは depth=3 で全パッケージに到達してしまうため 2 に抑える。
const defaultNeighborhoodDepth = 2

// PreparedSource はclone・ロード・変更パッケージ特定の結果をまとめたもの。
type PreparedSource struct {
	RepoRoot            string // リポジトリルート（GitHub API パスの基点）
	RootDir             string // go.mod のあるディレクトリ（パッケージロードの基点）
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

// Prepare は HEAD SHA を対象に PrepareAtSHA を呼ぶ薄いラッパー。
func (s *SourceTree) Prepare(ctx context.Context, pr domain.PRInfo, token string, changed []domain.ChangedFile) (*PreparedSource, error) {
	return s.PrepareAtSHA(ctx, pr, token, pr.HeadSHA, changed)
}

// PrepareAtSHA は指定 SHA を clone → fast load → 変更パッケージ特定 → 近傍フルロード する。
//
// 2フェーズロードにより、プロジェクト内の変更パッケージとその近傍のみを
// 型情報付きでロードし、外部ライブラリを含む全パッケージのロードを回避する。
//
// Loadが失敗したときはCloneのCleanupを内部で呼んでから返す（呼び出し側に部分状態を渡さない）。
// 成功時のみPreparedSource.Cleanupが返り、呼び出し側がdeferで呼ぶ責務を持つ。
// モノレポ対応: cloneルートでパッケージが見つからない場合、go.mod を持つサブディレクトリを自動検出する。
//
// changed の Filename は sha 側のリポジトリに存在するパスを指している前提（base 側を呼び出す
// ときに renamed/removed のパス補正が必要なら呼び出し側で行うこと）。
func (s *SourceTree) PrepareAtSHA(ctx context.Context, pr domain.PRInfo, token, sha string, changed []domain.ChangedFile) (*PreparedSource, error) {
	cloned, err := s.Cloner.Clone(ctx, CloneRequest{
		Owner: pr.Owner,
		Repo:  pr.Repo,
		SHA:   sha,
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("analyzer: clone %s/%s@%s: %w", pr.Owner, pr.Repo, sha, err)
	}

	ps, err := s.prepareFromDir(ctx, cloned.RootDir, changed)
	if err != nil {
		_ = cloned.Cleanup()
		return nil, err
	}
	ps.Cleanup = cloned.Cleanup
	return ps, nil
}

// CloneBoth は head/base の2SHAを1リポジトリ共有でcloneし、SHAごとのworktreeを返す。
// fetchとオブジェクトストアを共有するため、base/headを別々にcloneするより高速・省ディスク
// （変更のないblobは1度しかfetchしない）。返り値の Cleanup を呼び出し側がdeferで呼ぶ。
func (s *SourceTree) CloneBoth(ctx context.Context, pr domain.PRInfo, token string) (*ClonedWorktrees, error) {
	cw, err := s.Cloner.CloneWorktrees(ctx, WorktreesRequest{
		Owner: pr.Owner,
		Repo:  pr.Repo,
		Token: token,
		SHAs:  []string{pr.HeadSHA, pr.BaseSHA},
	})
	if err != nil {
		return nil, fmt.Errorf("analyzer: clone %s/%s (head=%s base=%s): %w", pr.Owner, pr.Repo, pr.HeadSHA, pr.BaseSHA, err)
	}
	return cw, nil
}

// PrepareFromDir はclone済みの作業ツリー（rootDir）に対して fast load →
// 変更パッケージ特定 → 近傍フルロードを行う。cloneを行わない点だけが PrepareAtSHA と
// 異なり、CloneBoth と組み合わせてcloneを共有する用途に使う。
// 返す PreparedSource の Cleanup は nil（共有tmp dirの後始末は CloneBoth 側の Cleanup が担う）。
func (s *SourceTree) PrepareFromDir(ctx context.Context, rootDir string, changed []domain.ChangedFile) (*PreparedSource, error) {
	return s.prepareFromDir(ctx, rootDir, changed)
}

// prepareFromDir はclone済みディレクトリを起点に2フェーズロードを実行する共通処理。
// clone/cleanupには関与しない（エラー時もcleanupを呼ばず、呼び出し側の責務とする）。
// 返す PreparedSource の Cleanup は常に nil。
func (s *SourceTree) prepareFromDir(ctx context.Context, rootDir string, changed []domain.ChangedFile) (*PreparedSource, error) {
	// go.mod を持つサブディレクトリを特定し、そこを基点にパッケージをロードする。
	// clonedRoot はリポジトリルート（GitHub API の相対パス基点）、loadDir は go.mod の親。
	clonedRoot := rootDir
	loadDir, err := findGoModRoot(clonedRoot)
	if err != nil {
		return nil, fmt.Errorf("analyzer: find go.mod in %s: %w", clonedRoot, err)
	}

	// Phase 1: 型チェックなしで全プロジェクトパッケージの構造を高速ロード。
	// import グラフのみ取得し、変更パッケージとその近傍を特定する。
	t0 := time.Now()
	fastPkgs, err := s.Loader.FastLoad(ctx, loadDir)
	if err != nil {
		return nil, fmt.Errorf("analyzer: fast-load packages at %s: %w", loadDir, err)
	}
	log.Printf("analyzer: phase1 FastLoad(%d pkgs) took %s", len(fastPkgs), time.Since(t0))

	// リポジトリ全体のパッケージ数が上限を超えるなら、Phase 2（型情報付きロード）に
	// 進む前に reject する。型ロードやジョブタイムアウトが無言で発火するのを防ぎ、
	// ユーザーに「大規模リポジトリは未対応」と明示するための入口側ガード。
	if len(fastPkgs) > maxFastLoadPackages {
		return nil, fmt.Errorf("%w: detected %d packages (limit %d)", ErrRepositoryTooLarge, len(fastPkgs), maxFastLoadPackages)
	}

	// GitHub API の changed ファイルパスはリポジトリルート相対なので clonedRoot を基点にする。
	changedPkgs := IdentifyChangedPackages(clonedRoot, fastPkgs, changed)
	log.Printf("analyzer: changed packages: %v", changedPkgs)

	// sha 側に実在するファイルの絶対パスを収集する。
	// HEAD 側では removed が、BASE 側では added が、それぞれ実在しないため自動的に除外される。
	changedFileAbsPaths := make([]string, 0, len(changed))
	for _, f := range changed {
		absPath := filepath.Clean(filepath.Join(clonedRoot, f.Filename))
		if _, statErr := os.Stat(absPath); statErr != nil {
			continue
		}
		changedFileAbsPaths = append(changedFileAbsPaths, absPath)
	}

	if len(changedPkgs) == 0 {
		// 変更されたGoパッケージがない（非Goファイルのみの変更など）
		return &PreparedSource{
			RepoRoot:            clonedRoot,
			RootDir:             loadDir,
			ChangedPackages:     nil,
			ChangedFileAbsPaths: changedFileAbsPaths,
		}, nil
	}

	// Phase 2: 変更パッケージとその近傍（defaultMaxDepth ホップ以内・最大
	// defaultMaxNeighborhood 件）のみを型情報付きでロードする。外部ライブラリは
	// 型チェックの依存として読まれるが返却パッケージには含まれない。
	// 件数を上限で抑えるのは、ハブパッケージを含む PR（k8s 等）で近傍が数千件に
	// 膨らみ、フルロード時にメモリ枯渇（OOM）するのを防ぐため。
	neighborhood := FindNeighborhood(changedPkgs, fastPkgs, defaultNeighborhoodDepth, defaultMaxNeighborhood)
	if len(neighborhood) >= defaultMaxNeighborhood {
		log.Printf("analyzer: neighborhood capped at %d pkgs (changed=%d) — graph may be partial", defaultMaxNeighborhood, len(changedPkgs))
	}
	log.Printf("analyzer: neighborhood(%d pkgs)", len(neighborhood))

	t1 := time.Now()
	result, err := s.Loader.Load(ctx, loadDir, neighborhood)
	if err != nil {
		return nil, fmt.Errorf("analyzer: load packages at %s: %w", loadDir, err)
	}
	log.Printf("analyzer: phase2 Load(%d pkgs) took %s", len(result.Packages), time.Since(t1))

	return &PreparedSource{
		RepoRoot:            clonedRoot,
		RootDir:             loadDir,
		Packages:            result.Packages,
		LoadErrors:          result.Errors,
		ChangedPackages:     changedPkgs,
		ChangedFileAbsPaths: changedFileAbsPaths,
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
