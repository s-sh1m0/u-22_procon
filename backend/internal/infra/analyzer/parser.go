package analyzer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// ErrNoPackages はGo パッケージが1件も見つからなかったときのsentinel error。
var ErrNoPackages = errors.New("analyzer: no Go packages found")

// loadMode は callgraph (#5) と diff→AST マップ (#6) で必要なフラグを全て含む。
const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedDeps |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes |
	packages.NeedModule

// fastLoadMode は型チェックなしでパッケージ構造だけを取得する高速フラグ。
// NeedImports まで含めることで import グラフから近傍パッケージを特定できる。
const fastLoadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedImports |
	packages.NeedModule

// LoadResult はパッケージロードの結果。
type LoadResult struct {
	// Packages はロードされたパッケージ。
	Packages []*packages.Package
	// Errors はパッケージ個別の型エラー等（致命エラーではない）。
	Errors []packages.Error
}

// PackageLoader はgo/packagesでパッケージ群をロードする抽象。
type PackageLoader interface {
	// FastLoad は型チェックなしでパッケージ構造だけを高速ロードする。
	// 変更パッケージ特定と近傍計算に使う。
	FastLoad(ctx context.Context, rootDir string) ([]*packages.Package, error)
	// Load は指定パターンのパッケージを型情報付きでロードする。
	// patterns はパッケージインポートパスのリスト。
	Load(ctx context.Context, rootDir string, patterns []string) (*LoadResult, error)
}

// GoPackageLoader はgo/packagesを使う本番実装。
type GoPackageLoader struct{}

// NewGoPackageLoader は本番用GoPackageLoaderを返す。
func NewGoPackageLoader() *GoPackageLoader {
	return &GoPackageLoader{}
}

// FastLoad はrootDir配下の全Goパッケージを型チェックなしで高速ロードする。
// 変更パッケージの特定と近傍パッケージの計算にのみ使う。
func (l *GoPackageLoader) FastLoad(ctx context.Context, rootDir string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Mode:    fastLoadMode,
		Dir:     rootDir,
		Context: ctx,
		Tests:   false,
		Env:     append(os.Environ(), "GOWORK=off"),
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoPackages, err)
	}

	var usable int
	for _, pkg := range pkgs {
		if len(pkg.GoFiles) > 0 {
			usable++
		}
	}
	if usable == 0 {
		return nil, ErrNoPackages
	}
	return pkgs, nil
}

// Load はpatterns で指定したパッケージを型情報付きでロードする。
// GOWORK=offを強制してホスト環境のgo.workの影響を受けないようにする。
func (l *GoPackageLoader) Load(ctx context.Context, rootDir string, patterns []string) (*LoadResult, error) {
	if len(patterns) == 0 {
		return &LoadResult{}, nil
	}

	cfg := &packages.Config{
		Mode:    loadMode,
		Dir:     rootDir,
		Context: ctx,
		Tests:   false,
		Env:     append(os.Environ(), "GOWORK=off"),
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoPackages, err)
	}

	// パッケージ個別のエラーを収集（型エラー等は致命扱いしない）
	var errs []packages.Error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		errs = append(errs, pkg.Errors...)
	})

	var usable int
	for _, pkg := range pkgs {
		if len(pkg.GoFiles) > 0 || len(pkg.CompiledGoFiles) > 0 {
			usable++
		}
	}
	if usable == 0 {
		return nil, ErrNoPackages
	}

	return &LoadResult{
		Packages: pkgs,
		Errors:   errs,
	}, nil
}

// IdentifyChangedPackages は変更ファイル（リポジトリルート相対パス）が属するパッケージのIDを返す。
// rootDirはEvalSymlinks適用済みのcloneディレクトリ絶対パス。
// 結果は重複なし・ソート済み。removed/renamedファイルは対象外（#6で対応予定）。
func IdentifyChangedPackages(rootDir string, pkgs []*packages.Package, changed []domain.ChangedFile) []string {
	// 変更ファイルの絶対パスをsetに入れる（added/modifiedのみ）
	wantAbs := make(map[string]struct{}, len(changed))
	for _, f := range changed {
		if f.Status == "removed" || f.Status == "renamed" {
			continue
		}
		abs := filepath.Clean(filepath.Join(rootDir, f.Filename))
		wantAbs[abs] = struct{}{}
	}
	if len(wantAbs) == 0 {
		return nil
	}

	rootPrefix := rootDir + string(filepath.Separator)

	seen := make(map[string]struct{})
	packages.Visit(pkgs, func(pkg *packages.Package) bool {
		for _, gf := range pkg.GoFiles {
			// rootDir配下のファイルのみ対象（cgo生成物など外部ファイルを除外）
			if !strings.HasPrefix(gf, rootPrefix) {
				continue
			}
			if _, ok := wantAbs[filepath.Clean(gf)]; ok {
				seen[pkg.ID] = struct{}{}
				break
			}
		}
		return true
	}, nil)

	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

// FindNeighborhood は変更パッケージを起点に import グラフを maxDepth ホップ辿った
// プロジェクト内パッケージのIDスライスを返す（重複なし・ソート済み）。
// allPkgs は FastLoad で得たプロジェクト内パッケージのみを想定する。
// 外部ライブラリは含まない。
func FindNeighborhood(changedPkgs []string, allPkgs []*packages.Package, maxDepth int) []string {
	// プロジェクトパッケージIDセット
	projectIDs := make(map[string]struct{}, len(allPkgs))
	for _, pkg := range allPkgs {
		projectIDs[pkg.ID] = struct{}{}
	}

	// 前方 import グラフ（callee方向: pkg → import先プロジェクトパッケージ）
	forward := make(map[string][]string, len(allPkgs))
	for _, pkg := range allPkgs {
		for impPath := range pkg.Imports {
			if _, ok := projectIDs[impPath]; ok {
				forward[pkg.ID] = append(forward[pkg.ID], impPath)
			}
		}
	}

	// 逆 import グラフ（caller方向: pkg → import元プロジェクトパッケージ）
	reverse := make(map[string][]string, len(allPkgs))
	for callerID, callees := range forward {
		for _, calleeID := range callees {
			reverse[calleeID] = append(reverse[calleeID], callerID)
		}
	}

	// BFS（双方向、maxDepth ホップ）
	type entry struct {
		id    string
		depth int
	}
	visited := make(map[string]struct{})
	queue := make([]entry, 0, len(changedPkgs))
	for _, id := range changedPkgs {
		if _, ok := projectIDs[id]; !ok {
			continue
		}
		if _, ok := visited[id]; !ok {
			visited[id] = struct{}{}
			queue = append(queue, entry{id, 0})
		}
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.depth >= maxDepth {
			continue
		}
		neighbors := append(forward[cur.id], reverse[cur.id]...)
		for _, nid := range neighbors {
			if _, ok := visited[nid]; !ok {
				visited[nid] = struct{}{}
				queue = append(queue, entry{nid, cur.depth + 1})
			}
		}
	}

	result := make([]string, 0, len(visited))
	for id := range visited {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}
