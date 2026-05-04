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
	result, err := loader.Load(context.Background(), dir)
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
	dir := t.TempDir() // 空ディレクトリ（go.modなし）

	loader := NewGoPackageLoader()
	_, err := loader.Load(context.Background(), dir)
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
	result, err := loader.Load(context.Background(), dir)
	// 型エラーがあっても Load 自体は成功して Errors に詰めて返す
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one type error in result.Errors, got none")
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
	result, err := loader.Load(context.Background(), dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
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
			got := IdentifyChangedPackages(dir, result.Packages, tc.changed)
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
