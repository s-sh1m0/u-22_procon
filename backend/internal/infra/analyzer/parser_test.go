package analyzer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// writeFixture はt.TempDir()内にGoモジュールfixureを書き出す。
// ファイルマップ: path（モジュールルート相対）→ content
func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(abs), err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", abs, err)
		}
	}
	return dir
}

func TestGoPackageLoader_FastLoad_Fixture(t *testing.T) {
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

	loader := NewGoPackageLoader()
	pkgs, err := loader.FastLoad(context.Background(), dir)
	if err != nil {
		t.Fatalf("FastLoad: %v", err)
	}

	pkgIDs := make(map[string]struct{})
	for _, pkg := range pkgs {
		pkgIDs[pkg.ID] = struct{}{}
	}

	for _, want := range []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/b"} {
		if _, ok := pkgIDs[want]; !ok {
			t.Errorf("package %q not found; got: %v", want, pkgIDs)
		}
	}
}

func TestGoPackageLoader_FastLoad_NoModule(t *testing.T) {
	dir := t.TempDir()

	loader := NewGoPackageLoader()
	_, err := loader.FastLoad(context.Background(), dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNoPackages) {
		t.Errorf("expected ErrNoPackages, got: %v", err)
	}
}

func TestGoPackageLoader_Load_Fixture(t *testing.T) {
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

	loader := NewGoPackageLoader()
	result, err := loader.Load(context.Background(), dir, []string{"./..."})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// パッケージIDのセットを確認
	pkgIDs := make(map[string]struct{})
	packages.Visit(result.Packages, func(pkg *packages.Package) bool {
		pkgIDs[pkg.ID] = struct{}{}
		return true
	}, nil)

	for _, want := range []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/b"} {
		if _, ok := pkgIDs[want]; !ok {
			t.Errorf("package %q not found in loaded packages; got: %v", want, pkgIDs)
		}
	}

	// 型エラーがないことを確認
	if len(result.Errors) != 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// TypesInfoが設定されていることを確認（callgraph構築に必要）
	packages.Visit(result.Packages, func(pkg *packages.Package) bool {
		if pkg.ID == "example.com/fixture/pkg/a" || pkg.ID == "example.com/fixture/pkg/b" {
			if pkg.TypesInfo == nil {
				t.Errorf("package %q has nil TypesInfo", pkg.ID)
			}
		}
		return true
	}, nil)
}

func TestGoPackageLoader_Load_NoModule(t *testing.T) {
	dir := t.TempDir()

	loader := NewGoPackageLoader()
	_, err := loader.Load(context.Background(), dir, []string{"./..."})
	if err == nil {
		t.Fatal("expected error for empty directory, got nil")
	}
	if !errors.Is(err, ErrNoPackages) {
		t.Errorf("expected ErrNoPackages, got: %v", err)
	}
}

func TestGoPackageLoader_Load_TypeError(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/broken\n\ngo 1.21\n",
		"main.go": `package main

func F() int { return "not an int" } // 型エラー
`,
	})

	loader := NewGoPackageLoader()
	result, err := loader.Load(context.Background(), dir, []string{"./..."})
	// 型エラーがあっても Load 自体は成功して Errors に詰めて返す
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one type error in result.Errors, got none")
	}
}

func TestGoPackageLoader_Load_EmptyPatterns(t *testing.T) {
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/fixture\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

func F() {}
`,
	})

	loader := NewGoPackageLoader()
	result, err := loader.Load(context.Background(), dir, []string{})
	if err != nil {
		t.Fatalf("Load with empty patterns: %v", err)
	}
	if len(result.Packages) != 0 {
		t.Errorf("expected empty packages, got %d", len(result.Packages))
	}
}

func TestIdentifyChangedPackages(t *testing.T) {
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

	loader := NewGoPackageLoader()
	// IdentifyChangedPackages は GoFiles 情報だけあれば動作するので FastLoad で足りる
	pkgs, err := loader.FastLoad(context.Background(), dir)
	if err != nil {
		t.Fatalf("FastLoad: %v", err)
	}

	tests := []struct {
		name    string
		changed []domain.ChangedFile
		want    []string
	}{
		{
			name:    "pkg/aのみ変更",
			changed: []domain.ChangedFile{{Filename: "pkg/a/a.go", Status: "modified"}},
			want:    []string{"example.com/fixture/pkg/a"},
		},
		{
			name: "pkg/aとpkg/bの両方変更",
			changed: []domain.ChangedFile{
				{Filename: "pkg/a/a.go", Status: "modified"},
				{Filename: "pkg/b/b.go", Status: "added"},
			},
			want: []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/b"},
		},
		{
			name:    "存在しないファイル",
			changed: []domain.ChangedFile{{Filename: "pkg/c/c.go", Status: "modified"}},
			want:    nil,
		},
		{
			name:    "removedは無視される",
			changed: []domain.ChangedFile{{Filename: "pkg/a/a.go", Status: "removed"}},
			want:    nil,
		},
		{
			name:    "renamedは無視される",
			changed: []domain.ChangedFile{{Filename: "pkg/a/a.go", Status: "renamed"}},
			want:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IdentifyChangedPackages(dir, pkgs, tc.changed)
			if len(got) != len(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
				return
			}
			for i, id := range got {
				if id != tc.want[i] {
					t.Errorf("[%d] got %q, want %q", i, id, tc.want[i])
				}
			}
		})
	}
}

func TestFindNeighborhood(t *testing.T) {
	// fixture: a → b → c という import チェーン、d は孤立
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/fixture\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

import "example.com/fixture/pkg/b"

func FA() { b.FB() }
`,
		"pkg/b/b.go": `package b

import "example.com/fixture/pkg/c"

func FB() { c.FC() }
`,
		"pkg/c/c.go": `package c

func FC() {}
`,
		"pkg/d/d.go": `package d

func FD() {}
`,
	})

	loader := NewGoPackageLoader()
	fastPkgs, err := loader.FastLoad(context.Background(), dir)
	if err != nil {
		t.Fatalf("FastLoad: %v", err)
	}

	tests := []struct {
		name         string
		changedPkgs  []string
		maxDepth     int
		wantIncludes []string
		wantExcludes []string
	}{
		{
			name:         "pkg/b変更・depth=1: aとcが近傍に入る",
			changedPkgs:  []string{"example.com/fixture/pkg/b"},
			maxDepth:     1,
			wantIncludes: []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/b", "example.com/fixture/pkg/c"},
			wantExcludes: []string{"example.com/fixture/pkg/d"},
		},
		{
			name:         "pkg/c変更・depth=1: bが近傍に入るがaは入らない",
			changedPkgs:  []string{"example.com/fixture/pkg/c"},
			maxDepth:     1,
			wantIncludes: []string{"example.com/fixture/pkg/b", "example.com/fixture/pkg/c"},
			wantExcludes: []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/d"},
		},
		{
			name:         "pkg/c変更・depth=2: aまで到達",
			changedPkgs:  []string{"example.com/fixture/pkg/c"},
			maxDepth:     2,
			wantIncludes: []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/b", "example.com/fixture/pkg/c"},
			wantExcludes: []string{"example.com/fixture/pkg/d"},
		},
		{
			name:         "孤立パッケージ変更: 自分のみ",
			changedPkgs:  []string{"example.com/fixture/pkg/d"},
			maxDepth:     3,
			wantIncludes: []string{"example.com/fixture/pkg/d"},
			wantExcludes: []string{"example.com/fixture/pkg/a", "example.com/fixture/pkg/b", "example.com/fixture/pkg/c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FindNeighborhood(tc.changedPkgs, fastPkgs, tc.maxDepth)
			gotSet := make(map[string]struct{}, len(got))
			for _, id := range got {
				gotSet[id] = struct{}{}
			}
			for _, want := range tc.wantIncludes {
				if _, ok := gotSet[want]; !ok {
					t.Errorf("expected %q in neighborhood, got: %v", want, got)
				}
			}
			for _, notWant := range tc.wantExcludes {
				if _, ok := gotSet[notWant]; ok {
					t.Errorf("unexpected %q in neighborhood, got: %v", notWant, got)
				}
			}
			// ソート済みであることを確認
			if !sort.StringsAreSorted(got) {
				t.Errorf("result is not sorted: %v", got)
			}
		})
	}
}
