package analyzer

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestClassifyLoadErr(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want error
	}{
		{
			name: "toolchain DL 失敗",
			in:   "err: exit status 1: stderr: go: download go1.99.0 for linux/amd64: toolchain not available",
			want: ErrGoToolchainUnavailable,
		},
		{
			name: "要求 Go バージョン超過",
			in:   "err: exit status 1: stderr: go: go.mod requires go >= 1.26.0 (running go 1.25.9)",
			want: ErrGoToolchainUnavailable,
		},
		{
			name: "go.mod 不在（Go リポジトリではない）",
			in:   "err: exit status 1: stderr: pattern ./...: directory prefix . does not contain main module or its selected dependencies",
			want: ErrNoPackages,
		},
		{
			name: "その他のロード失敗",
			in:   "err: exit status 1: stderr: could not import example.com/dep (invalid package name)",
			want: ErrPackageLoadFailed,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyLoadErr(errors.New(tc.in))
			if !errors.Is(got, tc.want) {
				t.Errorf("classifyLoadErr(%q) = %v; want errors.Is(_, %v)", tc.in, got, tc.want)
			}
			// 元のエラー文言が失われていないこと（デバッグ用に保持する）。
			if !strings.Contains(got.Error(), tc.in) {
				t.Errorf("classified error %q lost original message %q", got.Error(), tc.in)
			}
		})
	}
}

func TestGoPackageLoader_FastLoad_Workspace(t *testing.T) {
	// go.work で root と ./sub の2モジュールを束ねる。"./..." はモジュール境界で
	// 止まるため、workspacePatterns が両モジュールを列挙できることを確認する。
	dir := writeFixture(t, map[string]string{
		"go.work": "go 1.21\n\nuse (\n\t.\n\t./sub\n)\n",
		"go.mod":  "module example.com/wsfix\n\ngo 1.21\n",
		"root.go": `package wsfix

func Root() {}
`,
		"sub/go.mod": "module example.com/wsfix/sub\n\ngo 1.21\n",
		"sub/sub.go": `package sub

func Sub() {}
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

	for _, want := range []string{"example.com/wsfix", "example.com/wsfix/sub"} {
		if _, ok := pkgIDs[want]; !ok {
			t.Errorf("workspace module %q not found; got: %v", want, pkgIDs)
		}
	}
}

func TestFindNeighborhood_Cap(t *testing.T) {
	// a → b → c → d の import チェーン。a 起点・深さ十分でも maxPkgs=2 で打ち切られ、
	// 起点の a は必ず含まれることを確認する。
	dir := writeFixture(t, map[string]string{
		"go.mod": "module example.com/cap\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

import "example.com/cap/pkg/b"

func FA() { b.FB() }
`,
		"pkg/b/b.go": `package b

import "example.com/cap/pkg/c"

func FB() { c.FC() }
`,
		"pkg/c/c.go": `package c

import "example.com/cap/pkg/d"

func FC() { d.FD() }
`,
		"pkg/d/d.go": `package d

func FD() {}
`,
	})

	loader := NewGoPackageLoader()
	pkgs, err := loader.FastLoad(context.Background(), dir)
	if err != nil {
		t.Fatalf("FastLoad: %v", err)
	}

	got := FindNeighborhood([]string{"example.com/cap/pkg/a"}, pkgs, 5, 2)
	if len(got) > 2 {
		t.Errorf("neighborhood not capped: got %d (%v), want <= 2", len(got), got)
	}
	var hasA bool
	for _, id := range got {
		if id == "example.com/cap/pkg/a" {
			hasA = true
		}
	}
	if !hasA {
		t.Errorf("changed package a missing from capped neighborhood: %v", got)
	}
}
