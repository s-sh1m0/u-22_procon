package analyzer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// fakeCloner はテスト用のCloner実装。
// 指定したfixtureDirをそのまま使い、Cleanupの呼び出し有無を追跡する。
type fakeCloner struct {
	fixtureDir    string
	cleanupCalled bool
}

func (f *fakeCloner) Clone(_ context.Context, _ CloneRequest) (*ClonedRepo, error) {
	return &ClonedRepo{
		RootDir: f.fixtureDir,
		Cleanup: func() error {
			f.cleanupCalled = true
			return nil
		},
	}, nil
}

// errorCloner は常にエラーを返すCloner。
type errorCloner struct{}

func (e *errorCloner) Clone(_ context.Context, _ CloneRequest) (*ClonedRepo, error) {
	return nil, ErrCloneFailed
}

// errorLoader は常にエラーを返すPackageLoader。
type errorLoader struct{}

func (e *errorLoader) FastLoad(_ context.Context, _ string) ([]*packages.Package, error) {
	return nil, ErrNoPackages
}

func (e *errorLoader) Load(_ context.Context, _ string, _ []string) (*LoadResult, error) {
	return nil, ErrNoPackages
}

func TestSourceTree_Prepare_Success(t *testing.T) {
	// fixture: 2パッケージ構成のGoモジュール
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/fixture\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

import "example.com/fixture/pkg/b"

func F() int { return b.G() }
`,
		"pkg/b/b.go": `package b

func G() int { return 42 }
`,
	})

	// EvalSymlinksでパス正規化（SourceTree.Prepareの動作に合わせる）
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	fc := &fakeCloner{fixtureDir: resolvedDir}
	st := NewSourceTree(fc, NewGoPackageLoader())

	pr := domain.PRInfo{Owner: "owner", Repo: "repo", HeadSHA: "abc123"}
	changed := []domain.ChangedFile{
		{Filename: "pkg/a/a.go", Status: "modified"},
		{Filename: "pkg/b/b.go", Status: "added"},
	}

	ps, err := st.Prepare(context.Background(), pr, "", changed)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	if ps.RepoRoot != resolvedDir {
		t.Errorf("RepoRoot: got %q, want %q", ps.RepoRoot, resolvedDir)
	}
	if ps.RootDir != resolvedDir {
		t.Errorf("RootDir: got %q, want %q", ps.RootDir, resolvedDir)
	}
	if len(ps.Packages) == 0 {
		t.Error("Packages should not be empty")
	}
	if len(ps.ChangedPackages) != 2 {
		t.Errorf("ChangedPackages: got %v, want 2 entries", ps.ChangedPackages)
	}

	// Cleanup でdirが（fakeなので実際には消えないが）呼ばれることを確認
	if err := ps.Cleanup(); err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if !fc.cleanupCalled {
		t.Error("expected fakeCloner.Cleanup to be called")
	}
}

func TestSourceTree_Prepare_MonorepoRepoRoot(t *testing.T) {
	// fixture: go.mod がサブディレクトリ backend/ にあるモノレポ構成
	dir := writeFixture(t, map[string]string{
		"backend/go.mod": "module example.com/fixture\n\ngo 1.21\n",
		"backend/pkg/a/a.go": `package a

func F() int { return 42 }
`,
	})

	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	fc := &fakeCloner{fixtureDir: resolvedDir}
	st := NewSourceTree(fc, NewGoPackageLoader())

	pr := domain.PRInfo{Owner: "owner", Repo: "repo", HeadSHA: "abc123"}
	changed := []domain.ChangedFile{
		{Filename: "backend/pkg/a/a.go", Status: "modified"},
	}

	ps, err := st.Prepare(context.Background(), pr, "", changed)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	defer func() { _ = ps.Cleanup() }()

	wantRepoRoot := resolvedDir
	wantRootDir := filepath.Join(resolvedDir, "backend")

	if ps.RepoRoot != wantRepoRoot {
		t.Errorf("RepoRoot: got %q, want %q", ps.RepoRoot, wantRepoRoot)
	}
	if ps.RootDir != wantRootDir {
		t.Errorf("RootDir: got %q, want %q", ps.RootDir, wantRootDir)
	}
}

func TestSourceTree_Prepare_CloneError(t *testing.T) {
	st := NewSourceTree(&errorCloner{}, NewGoPackageLoader())

	pr := domain.PRInfo{Owner: "o", Repo: "r", HeadSHA: "sha"}
	_, err := st.Prepare(context.Background(), pr, "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrCloneFailed) {
		t.Errorf("expected ErrCloneFailed, got: %v", err)
	}
}

func TestSourceTree_Prepare_LoaderFailure_CleansUp(t *testing.T) {
	dir := t.TempDir() // 空ディレクトリ（go.modなし） → ErrNoPackages

	fc := &fakeCloner{fixtureDir: dir}
	st := NewSourceTree(fc, &errorLoader{})

	pr := domain.PRInfo{Owner: "o", Repo: "r", HeadSHA: "sha"}
	_, err := st.Prepare(context.Background(), pr, "", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNoPackages) {
		t.Errorf("expected ErrNoPackages, got: %v", err)
	}
	// Loaderが失敗したときにCleanupが呼ばれていることを確認（部分状態を残さない契約）
	if !fc.cleanupCalled {
		t.Error("expected fakeCloner.Cleanup to be called on Loader failure")
	}
	// 実際のdirがまだ存在することを確認（fakeなので消されない）
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("fixture dir should still exist: %v", err)
	}
}
