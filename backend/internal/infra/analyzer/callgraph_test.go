package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// setupCGFixture は callgraph テスト用の Go モジュールを tmpdir に構築して返す。
//
// モジュール構成:
//
//	example.com/cgfixture
//	  pkg/a/a.go  -- func A() { b.B() }  /  func C() {}
//	  pkg/b/b.go  -- func B() {}
func setupCGFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeFile := func(rel, content string) {
		t.Helper()
		abs := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeFile("go.mod", "module example.com/cgfixture\n\ngo 1.21\n")
	writeFile("pkg/a/a.go", `package a

import "example.com/cgfixture/pkg/b"

func A() { b.B() }
func C() {}
`)
	writeFile("pkg/b/b.go", `package b

func B() {}
`)
	return dir
}

// loadCGFixture は setupCGFixture で構築したモジュールを packages.Load でロードして返す。
func loadCGFixture(t *testing.T, rootDir string) []*packages.Package {
	t.Helper()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedModule,
		Dir:   rootDir,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}
	return pkgs
}

func TestGoCallGraphBuilder_Build_ChangedNodes(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// 変更パッケージ内の関数（A, C）は Changed=true
	// 呼び出し先（B）は Changed=false だがグラフに含まれる
	nodeMap := make(map[string]bool) // Name -> Changed
	for _, n := range graph.Nodes {
		nodeMap[n.Name] = n.Changed
	}

	tests := []struct {
		name    string
		changed bool
	}{
		{"A", true},
		{"C", true},
		{"B", false},
	}
	for _, tc := range tests {
		changed, ok := nodeMap[tc.name]
		if !ok {
			t.Errorf("node %q not found in graph (nodes: %v)", tc.name, nodeMap)
			continue
		}
		if changed != tc.changed {
			t.Errorf("node %q Changed=%v, want %v", tc.name, changed, tc.changed)
		}
	}
}

func TestGoCallGraphBuilder_Build_Edges(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// A → B のエッジが存在する
	nodeByName := make(map[string]string) // Name -> ID
	for _, n := range graph.Nodes {
		nodeByName[n.Name] = string(n.ID)
	}

	aID, ok := nodeByName["A"]
	if !ok {
		t.Fatal("node A not found")
	}
	bID, ok := nodeByName["B"]
	if !ok {
		t.Fatal("node B not found")
	}

	found := false
	for _, e := range graph.Edges {
		if string(e.From) == aID && string(e.To) == bID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("edge A->B not found (edges: %v)", graph.Edges)
	}
}

func TestGoCallGraphBuilder_Build_StdlibExcluded(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// stdlib のパッケージ（builtin, fmt 等）がノードに含まれない
	for _, n := range graph.Nodes {
		if n.Package == "fmt" || n.Package == "builtin" || n.Package == "runtime" {
			t.Errorf("stdlib package %q should not be in graph, but got node %q", n.Package, n.Name)
		}
	}
}

func TestGoCallGraphBuilder_Build_EmptyChanged(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(graph.Nodes) != 0 || len(graph.Edges) != 0 {
		t.Errorf("expected empty graph, got nodes=%d edges=%d", len(graph.Nodes), len(graph.Edges))
	}
}
