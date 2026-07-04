package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// setupCGFixture は callgraph テスト用の Go モジュールを tmpdir に構築して返す。
//
// モジュール構成:
//
//	example.com/cgfixture
//	  pkg/a/a.go  -- func A() { b.B() }  /  func C() {}  /  func D() { closure → C() }
//	                 type T1/T2 それぞれに func M() { closure → C() }
//	  pkg/b/b.go  -- func B() {}
//
// D はクロージャを内部で呼び出すため、SSA 上では D$1 という別関数になる。
// クロージャ畳み込みの検証に使う。
// T1.M / T2.M は同名メソッドで、それぞれ内部にクロージャを持つ。
// NodeID にレシーバ型が含まれず衝突しないこと（およびクロージャ畳み込みでも
// 衝突しないこと）の検証に使う。
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

func D() {
	f := func() { C() }
	f()
}

type T1 struct{}

func (T1) M() {
	f := func() { C() }
	f()
}

type T2 struct{}

func (T2) M() {
	f := func() { C() }
	f()
}
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

	// pkg/a/a.go だけを変更ファイルとして渡す（ファイル単位の changed フラグを検証）
	changedFileAbsPaths := []string{filepath.Join(root, "pkg/a/a.go")}

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, changedFileAbsPaths, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// File フィールドが相対パスになっている（絶対パスでない）
	for _, n := range graph.Nodes {
		if filepath.IsAbs(n.File) {
			t.Errorf("node %q: File should be relative, got %q", n.Name, n.File)
		}
	}

	// 変更ファイル内の関数（A, C）は Changed=true
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
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, []string{filepath.Join(root, "pkg/a/a.go")}, root)
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
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, []string{filepath.Join(root, "pkg/a/a.go")}, root)
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

func TestGoCallGraphBuilder_Build_ClosureMerged(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, []string{filepath.Join(root, "pkg/a/a.go")}, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	nodeByName := make(map[string]string) // Name -> ID
	for _, n := range graph.Nodes {
		// クロージャ由来の `$連番` ノードが残っていないこと
		if strings.Contains(n.Name, "$") {
			t.Errorf("closure node should be merged into parent, but found %q", n.Name)
		}
		nodeByName[n.Name] = string(n.ID)
	}

	// 親関数 D のノードは存在する
	dID, ok := nodeByName["D"]
	if !ok {
		t.Fatalf("node D not found (nodes: %v)", nodeByName)
	}
	cID, ok := nodeByName["C"]
	if !ok {
		t.Fatalf("node C not found (nodes: %v)", nodeByName)
	}

	// クロージャが持っていた D$1 → C のエッジが親 D → C に付け替わっている
	foundDC := false
	for _, e := range graph.Edges {
		// 親→クロージャ由来の自己ループ D → D が残っていないこと
		if string(e.From) == dID && string(e.To) == dID {
			t.Errorf("self-loop D->D should be removed, but found %v", e)
		}
		if string(e.From) == dID && string(e.To) == cID {
			foundDC = true
		}
	}
	if !foundDC {
		t.Errorf("edge D->C (reattached from closure) not found (edges: %v)", graph.Edges)
	}
}

func TestGoCallGraphBuilder_Build_SameMethodNameNotCollided(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, []string{filepath.Join(root, "pkg/a/a.go")}, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// 同名メソッド T1.M / T2.M が NodeID 衝突せず別ノードとして存在する
	var mIDs []string
	var cID string
	for _, n := range graph.Nodes {
		switch n.Name {
		case "M":
			mIDs = append(mIDs, string(n.ID))
		case "C":
			cID = string(n.ID)
		}
	}
	if len(mIDs) != 2 {
		t.Fatalf("want 2 distinct nodes named M (T1.M, T2.M), got %d (nodes: %v)", len(mIDs), graph.Nodes)
	}
	if mIDs[0] == mIDs[1] {
		t.Errorf("T1.M and T2.M must have distinct NodeIDs, both = %q", mIDs[0])
	}

	// 各メソッド内クロージャの C 呼び出しが、衝突せずそれぞれの親メソッドに付け替わる
	for _, mID := range mIDs {
		found := false
		for _, e := range graph.Edges {
			if string(e.From) == mID && string(e.To) == cID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("edge %s->C not found (edges: %v)", mID, graph.Edges)
		}
	}
}

// setupCGFixtureExtDep は外部依存（stdlib）を含む Go モジュールを構築する。
// export data 経路（NeedDeps なし + ssautil.Packages）で外部パッケージの本体 SSA を
// 構築しなくても、プロジェクト内グラフが旧経路と一致することを検証するための fixture。
//
//	example.com/extfixture
//	  pkg/a/a.go -- func A() string { return strings.ToUpper(b.B()) }  ← 外部(strings)+内部(b.B)呼び出し
//	  pkg/b/b.go -- func B() string { return strings.TrimSpace("x") }  ← 外部(strings)呼び出し
func setupCGFixtureExtDep(t *testing.T) string {
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

	writeFile("go.mod", "module example.com/extfixture\n\ngo 1.21\n")
	writeFile("pkg/a/a.go", `package a

import (
	"strings"

	"example.com/extfixture/pkg/b"
)

func A() string { return strings.ToUpper(b.B()) }
`)
	writeFile("pkg/b/b.go", `package b

import "strings"

func B() string { return strings.TrimSpace(" x ") }
`)
	return dir
}

// TestGoCallGraphBuilder_Build_ExportDataParity は PR-F（依存を export data 化）の
// 計測ゲートに相当する回帰テスト。外部依存を持つ fixture を、
//   - 既定（loadMode に NeedDeps なし + ssautil.Packages）
//   - escape hatch（ANALYZER_SOURCE_DEPS=1 → NeedDeps + ssautil.AllPackages）
//
// の両経路で本番ローダ経由でロード→ Build し、プロジェクト内グラフのノード集合・
// エッジ集合が完全一致することを確認する。依存の SSA 本体を構築しなくても出力が
// 変わらないという PR-F の前提を CI で守る。
func TestGoCallGraphBuilder_Build_ExportDataParity(t *testing.T) {
	root := setupCGFixtureExtDep(t)
	changedPkgs := []string{
		"example.com/extfixture/pkg/a",
		"example.com/extfixture/pkg/b",
	}
	changedFiles := []string{
		filepath.Join(root, "pkg/a/a.go"),
		filepath.Join(root, "pkg/b/b.go"),
	}

	build := func(sourceDeps bool) (map[string]struct{}, map[[2]string]struct{}) {
		if sourceDeps {
			t.Setenv("ANALYZER_SOURCE_DEPS", "1")
		} else {
			t.Setenv("ANALYZER_SOURCE_DEPS", "0")
		}

		res, err := NewGoPackageLoader().Load(context.Background(), root, changedPkgs)
		if err != nil {
			t.Fatalf("Load(sourceDeps=%v): %v", sourceDeps, err)
		}
		b := &GoCallGraphBuilder{MaxDepth: 3}
		g, err := b.Build(context.Background(), res.Packages, changedPkgs, changedFiles, root)
		if err != nil {
			t.Fatalf("Build(sourceDeps=%v): %v", sourceDeps, err)
		}

		nodes := make(map[string]struct{}, len(g.Nodes))
		for _, n := range g.Nodes {
			// stdlib のパッケージがグラフに漏れていないことも同時に確認する。
			if n.Package == "strings" || n.Package == "fmt" || n.Package == "runtime" {
				t.Errorf("stdlib package %q leaked into graph (node %q, sourceDeps=%v)", n.Package, n.Name, sourceDeps)
			}
			nodes[string(n.ID)] = struct{}{}
		}
		edges := make(map[[2]string]struct{}, len(g.Edges))
		for _, e := range g.Edges {
			edges[[2]string{string(e.From), string(e.To)}] = struct{}{}
		}
		return nodes, edges
	}

	fastNodes, fastEdges := build(false)
	slowNodes, slowEdges := build(true)

	if len(fastNodes) == 0 {
		t.Fatal("expected non-empty graph on export-data path")
	}
	if !reflect.DeepEqual(fastNodes, slowNodes) {
		t.Errorf("node sets diverge between export-data and source-deps paths:\n export-data=%v\n source-deps=%v", fastNodes, slowNodes)
	}
	if !reflect.DeepEqual(fastEdges, slowEdges) {
		t.Errorf("edge sets diverge between export-data and source-deps paths:\n export-data=%v\n source-deps=%v", fastEdges, slowEdges)
	}
}

func TestGoCallGraphBuilder_Build_EmptyChanged(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{}, []string{}, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(graph.Nodes) != 0 || len(graph.Edges) != 0 {
		t.Errorf("expected empty graph, got nodes=%d edges=%d", len(graph.Nodes), len(graph.Edges))
	}
}
