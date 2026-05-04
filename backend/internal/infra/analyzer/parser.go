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

// LoadResult はパッケージロードの結果。
type LoadResult struct {
	// Packages はtransitivelyロードされた全パッケージ。
	Packages []*packages.Package
	// Errors はパッケージ個別の型エラー等（致命エラーではない）。
	Errors []packages.Error
}

// PackageLoader はgo/packagesでディレクトリ配下のパッケージ群をロードする抽象。
type PackageLoader interface {
	Load(ctx context.Context, rootDir string) (*LoadResult, error)
}

// GoPackageLoader はgo/packagesを使う本番実装。
type GoPackageLoader struct{}

// NewGoPackageLoader は本番用GoPackageLoaderを返す。
func NewGoPackageLoader() *GoPackageLoader {
	return &GoPackageLoader{}
}

// Load はrootDir配下の全Goパッケージを型情報付きでロードする。
// GOWORK=offを強制してホスト環境のgo.workの影響を受けないようにする。
func (l *GoPackageLoader) Load(ctx context.Context, rootDir string) (*LoadResult, error) {
	cfg := &packages.Config{
		Mode:    loadMode,
		Dir:     rootDir,
		Context: ctx,
		Tests:   false,
		// GOWORK=off: 開発機の外側go.workがcloneされたモジュールの解析に干渉しないよう強制
		Env: append(os.Environ(), "GOWORK=off"),
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNoPackages, err)
	}

	// パッケージ個別のエラーを収集（型エラー等は致命扱いしない）
	var errs []packages.Error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		errs = append(errs, pkg.Errors...)
	})

	// GoFilesが1件もないならGoパッケージが存在しないと判断する。
	// packages.Loadはgo.modが無い空ディレクトリでも0件ではなくエラー付きエントリを返すことがあるため。
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
