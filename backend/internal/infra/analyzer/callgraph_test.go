package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
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

	// func A() は 5 行目。patch は A の行のみ変更。
	// C (6行目)、D (8-11行目) は変更なし。
	aFile := filepath.Join(root, "pkg/a/a.go")
	cl := ChangedLines{
		aFile: {
			{Start: 5, End: 5},
		},
	}

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// File フィールドが相対パスになっている（絶対パスでない）
	for _, n := range graph.Nodes {
		if filepath.IsAbs(n.File) {
			t.Errorf("node %q: File should be relative, got %q", n.Name, n.File)
		}
	}

	// A のみ Changed=true、C / D / B は Changed=false
	nodeMap := make(map[string]bool) // Name -> Changed
	for _, n := range graph.Nodes {
		nodeMap[n.Name] = n.Changed
	}

	tests := []struct {
		name    string
		changed bool
	}{
		{"A", true},
		{"C", false},
		{"D", false},
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

func TestGoCallGraphBuilder_Build_LineRange(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	// a.go を変更ファイルにする（行範囲は空 = 変更関数なし）。起点は a.go の関数になる。
	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): {}}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	type lineRange struct{ line, endLine int }
	got := make(map[string]lineRange) // Name -> 行範囲
	for _, n := range graph.Nodes {
		got[n.Name] = lineRange{n.Line, n.EndLine}
	}

	// D はクロージャを畳み込んだ親関数の本体全体（8〜11 行目）を範囲に持つ。
	want := map[string]lineRange{
		"A": {5, 5},
		"C": {6, 6},
		"D": {8, 11},
		"B": {3, 3},
	}
	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("node %q not found in graph", name)
			continue
		}
		if g != w {
			t.Errorf("node %q Line/EndLine = %d/%d, want %d/%d", name, g.line, g.endLine, w.line, w.endLine)
		}
	}
}

func TestGoCallGraphBuilder_Build_ChangedNodes_FileFallback(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	// Patch が空（nil ranges）の場合はファイル単位にフォールバック → 全関数 changed
	aFile := filepath.Join(root, "pkg/a/a.go")
	cl := ChangedLines{
		aFile: nil,
	}

	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for _, n := range graph.Nodes {
		if n.Package == "example.com/cgfixture/pkg/a" && n.File != "" && !n.Changed {
			t.Errorf("node %q in changed file should be Changed=true with nil ranges (fallback), got false", n.Name)
		}
	}
}

func TestGoCallGraphBuilder_Build_Edges(t *testing.T) {
	root := setupCGFixture(t)
	pkgs := loadCGFixture(t, root)

	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): nil}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
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

	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): nil}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
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

	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): nil}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
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

	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): nil}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/cgfixture/pkg/a"}, cl, root)
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

// setupModule は files（モジュールルートからの相対パス → 内容）を書き出した一時ディレクトリを返す。
func setupModule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// nodeNames はグラフのノード名の集合を返す。
func nodeNames(g *domain.Graph) map[string]struct{} {
	names := make(map[string]struct{}, len(g.Nodes))
	for _, n := range g.Nodes {
		names[n.Name] = struct{}{}
	}
	return names
}

// TestGoCallGraphBuilder_Build_Scope は収集範囲が「変更ファイルの関数」を起点に、
// 呼び出し先・呼び出し元の各方向へ一方向にだけ広がることを確認する。
//
//	pkg/a/a.go     -- Changed() { b.Callee() }       ← 変更ファイル（起点）
//	pkg/a/other.go -- Unrelated(x int)               ← 同じパッケージだが変更ファイル外・依存なし
//	pkg/b/b.go     -- Callee() / Sibling() { Callee() }  ← Sibling は呼び出し先の別の呼び出し元
//	pkg/c/c.go     -- Caller() { a.Changed() } / CallerOfCaller() { Caller() }
func TestGoCallGraphBuilder_Build_Scope(t *testing.T) {
	root := setupModule(t, map[string]string{
		"go.mod": "module example.com/scope\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

import "example.com/scope/pkg/b"

func Changed() { b.Callee() }
`,
		"pkg/a/other.go": `package a

func Unrelated(x int) {}
`,
		"pkg/b/b.go": `package b

func Callee() {}

func Sibling() { Callee() }
`,
		"pkg/c/c.go": `package c

import "example.com/scope/pkg/a"

func Caller() { a.Changed() }

func CallerOfCaller() { Caller() }
`,
	})
	pkgs := loadCGFixture(t, root)

	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): {{Start: 5, End: 5}}}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/scope/pkg/a"}, cl, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	names := nodeNames(graph)
	tests := []struct {
		name string
		want bool
	}{
		{"Changed", true},        // 起点
		{"Callee", true},         // 呼び出し先
		{"Caller", true},         // 呼び出し元
		{"CallerOfCaller", true}, // 呼び出し元の呼び出し元（一方向に 2 ホップ）
		{"Sibling", false},       // 呼び出し先の別の呼び出し元（ジグザグ）
		{"Unrelated", false},     // 同じパッケージだが変更ファイル外・依存なし
	}
	for _, tc := range tests {
		if _, ok := names[tc.name]; ok != tc.want {
			t.Errorf("node %q in graph = %v, want %v (nodes: %v)", tc.name, ok, tc.want, names)
		}
	}
}

// TestGoCallGraphBuilder_Build_FuncValueCall は関数値の呼び出しが、同じシグネチャを持つ
// 無関係な関数へのエッジにならないことを確認する。インターフェース経由の呼び出しは
// 実装メソッドへのエッジとして残す。
//
//	pkg/a/a.go -- Run(fn func()) { fn() } / Use(i I) { i.M() }  ← 変更ファイル（起点）
//	pkg/b/b.go -- Unrelated() / T.M()                           ← Unrelated は fn と同じシグネチャ
func TestGoCallGraphBuilder_Build_FuncValueCall(t *testing.T) {
	root := setupModule(t, map[string]string{
		"go.mod": "module example.com/funcvalue\n\ngo 1.21\n",
		"pkg/a/a.go": `package a

type I interface{ M() }

func Run(fn func()) { fn() }

func Use(i I) { i.M() }
`,
		"pkg/b/b.go": `package b

func Unrelated() {}

type T struct{}

func (T) M() {}
`,
	})
	pkgs := loadCGFixture(t, root)

	cl := ChangedLines{filepath.Join(root, "pkg/a/a.go"): nil}
	b := &GoCallGraphBuilder{MaxDepth: 3}
	graph, err := b.Build(context.Background(), pkgs, []string{"example.com/funcvalue/pkg/a"}, cl, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	names := nodeNames(graph)
	if _, ok := names["Unrelated"]; ok {
		t.Errorf("Unrelated should not be reached via func value call fn() (nodes: %v)", names)
	}

	found := false
	for _, e := range graph.Edges {
		if e.From == "example.com/funcvalue/pkg/a.Use" && e.To == "(example.com/funcvalue/pkg/b.T).M" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("interface call edge Use -> T.M not found (edges: %v)", graph.Edges)
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
	changedLines := ChangedLines{
		filepath.Join(root, "pkg/a/a.go"): nil,
		filepath.Join(root, "pkg/b/b.go"): nil,
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
		g, err := b.Build(context.Background(), res.Packages, changedPkgs, changedLines, root)
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
	graph, err := b.Build(context.Background(), pkgs, []string{}, ChangedLines{}, root)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(graph.Nodes) != 0 || len(graph.Edges) != 0 {
		t.Errorf("expected empty graph, got nodes=%d edges=%d", len(graph.Nodes), len(graph.Edges))
	}
}
