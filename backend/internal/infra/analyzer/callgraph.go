package analyzer

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// defaultMaxDepth は起点から呼び出し先・呼び出し元の各方向に辿るホップ数のデフォルト上限。
const defaultMaxDepth = 3

// CallGraphBuilder はロード済みパッケージから呼び出しグラフを構築する抽象。
// source_tree.go が返す PreparedSource.Packages / ChangedPackages / ChangedLines をそのまま渡せる設計にする。
type CallGraphBuilder interface {
	Build(ctx context.Context, pkgs []*packages.Package, changedPkgIDs []string, changedLines ChangedLines, rootDir string) (*domain.Graph, error)
}

// GoCallGraphBuilder は golang.org/x/tools/go/callgraph を使う本番実装。
type GoCallGraphBuilder struct {
	// MaxDepth は起点から呼び出し先・呼び出し元の各方向に辿るホップ数の上限。0 以下でデフォルト値を使う。
	MaxDepth int
}

// NewGoCallGraphBuilder は GoCallGraphBuilder を返す。
func NewGoCallGraphBuilder() *GoCallGraphBuilder {
	return &GoCallGraphBuilder{MaxDepth: defaultMaxDepth}
}

// Build は pkgs から SSA を構築し、変更ファイルに定義された関数を起点に caller/callee グラフを返す。
// changedPkgIDs が空の場合は空の Graph を返す。
// changedLines は起点（変更ファイル）の判定と、changed フラグを関数単位で設定するために使う。
// stdlib・外部ライブラリはフィルタリングして含めない。
// rootDir はリポジトリのルートディレクトリの絶対パスで、ノードのファイルパスを相対パスに変換するために使用される。
func (b *GoCallGraphBuilder) Build(_ context.Context, pkgs []*packages.Package, changedPkgIDs []string, changedLines ChangedLines, rootDir string) (*domain.Graph, error) {
	if len(changedPkgIDs) == 0 {
		return &domain.Graph{}, nil
	}

	maxDepth := b.MaxDepth
	if maxDepth <= 0 {
		maxDepth = defaultMaxDepth
	}

	// 1. プロジェクト内パッケージのみを抽出する。
	// pkgs に外部ライブラリが含まれることがあるが（NeedDeps 有効時、または
	// 変更パッケージ自体が依存として現れる場合）、SSA 構築と loadedSet は
	// プロジェクト内パッケージのみを対象にする。
	// pkg.Module.Main が true のパッケージが自リポジトリのパッケージ。
	projectPkgs := make([]*packages.Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Main {
			projectPkgs = append(projectPkgs, pkg)
		}
	}

	// 2. SSA プログラムを構築（プロジェクト内パッケージのみ）。
	// ssautil.Packages は projectPkgs の関数本体のみ SSA 化し、依存パッケージの
	// 本体は構築しない。CHA が対象にするのは本体を持つ関数だが、出力グラフは
	// isInLoadedSet でプロジェクト内→内のエッジに限定するため、依存本体が無くても
	// 結果は変わらない。escape hatch が立つときのみ依存も含む AllPackages を使う。
	tSSA := time.Now()
	var prog *ssa.Program
	if sourceDepsEnabled() {
		prog, _ = ssautil.AllPackages(projectPkgs, ssa.InstantiateGenerics)
	} else {
		prog, _ = ssautil.Packages(projectPkgs, ssa.InstantiateGenerics)
	}
	prog.Build()
	ssaTook := time.Since(tSSA)

	// 3. CHA でコールグラフ構築（エントリポイント不要 → ライブラリにも適用可）
	tCHA := time.Now()
	cg := cha.CallGraph(prog)
	cg.DeleteSyntheticNodes()
	chaTook := time.Since(tCHA)

	// 4. 変更パッケージ ID セット（起点判定で Position を引く前の事前フィルタ）
	changedSet := make(map[string]struct{}, len(changedPkgIDs))
	for _, id := range changedPkgIDs {
		changedSet[id] = struct{}{}
	}

	// 5. プロジェクト内パッケージ ID セット（外部ライブラリ除外用）
	loadedSet := make(map[string]struct{}, len(projectPkgs))
	for _, pkg := range projectPkgs {
		loadedSet[pkg.ID] = struct{}{}
	}

	// 5. 変更ファイルに定義された関数を起点に、呼び出し先方向・呼び出し元方向へ
	// それぞれ maxDepth ホップ辿る。起点を関数単位にしないのは、base 側の changedLines が
	// head 側の行番号で作られていて base で削除された関数を行範囲から特定できず、
	// 中身の比較（RefineChangedByContent）も Build の後でしか行えないため。
	// 変更・追加・削除された関数は必ず変更ファイルにあるので、ファイル単位でも漏れない。
	// 各方向は一方向にだけ辿る（呼び出し先の別の呼び出し元、のような兄弟関数は拾わない）。
	tBFS := time.Now()
	var seeds []*callgraph.Node
	for fn, node := range cg.Nodes {
		if fn == nil || fn.Package() == nil {
			continue
		}
		if _, ok := changedSet[fn.Package().Pkg.Path()]; !ok {
			continue
		}
		if changedLines.FileChanged(prog.Fset.Position(fn.Pos()).Filename) {
			seeds = append(seeds, node)
		}
	}

	collected := make(map[*callgraph.Node]struct{})
	walk(seeds, true, maxDepth, loadedSet, collected)
	walk(seeds, false, maxDepth, loadedSet, collected)

	bfsTook := time.Since(tBFS)

	// 6. 収集ノード・エッジを domain.Graph に変換
	// 無名関数（クロージャ）は囲うトップレベル親関数に畳み込む。複数の callgraph
	// ノード（親本体 + 各クロージャ）が同一 NodeID を指すため、ノードは重複排除し、
	// エッジは親 NodeID へ付け替えたうえで自己ループ・重複を除去する。
	tConv := time.Now()
	nodeIDMap := make(map[*callgraph.Node]domain.NodeID, len(collected))
	nodeSeen := make(map[domain.NodeID]struct{}, len(collected))
	var nodes []domain.Node
	for cgNode := range collected {
		fn := rootFunc(cgNode.Func)
		nid := toNodeID(fn)
		nodeIDMap[cgNode] = nid
		if _, dup := nodeSeen[nid]; dup {
			continue
		}
		nodeSeen[nid] = struct{}{}

		pos := prog.Fset.Position(fn.Pos())
		funcChanged := false
		endLine := 0
		if syntaxNode := fn.Syntax(); syntaxNode != nil {
			endLine = prog.Fset.Position(syntaxNode.End()).Line
			funcChanged = changedLines.FunctionChanged(pos.Filename, pos.Line, endLine)
		} else {
			funcChanged = changedLines.FileChanged(pos.Filename)
		}
		relPath, err := filepath.Rel(rootDir, pos.Filename)
		if err != nil {
			relPath = pos.Filename
		}
		nodes = append(nodes, domain.Node{
			ID:         nid,
			Name:       fn.Name(),
			Package:    pkgPath(fn),
			File:       relPath,
			Line:       pos.Line,
			EndLine:    endLine,
			Changed:    funcChanged,
			DiffStatus: domain.DiffStatusExisting,
		})
	}

	var edges []domain.Edge
	seen := make(map[[2]domain.NodeID]struct{})
	for cgNode := range collected {
		callerID, ok := nodeIDMap[cgNode]
		if !ok {
			continue
		}
		for _, edge := range cgNode.Out {
			calleeID, ok2 := nodeIDMap[edge.Callee]
			if !ok2 {
				continue
			}
			if callerID == calleeID { // 親→クロージャ / クロージャ→クロージャ由来の自己ループを除去
				continue
			}
			key := [2]domain.NodeID{callerID, calleeID}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, domain.Edge{From: callerID, To: calleeID, Status: domain.DiffStatusExisting})
		}
	}

	// 内訳計時ログ: #82 のボトルネック計測用。SSA/CHA が支配的か BFS/変換が
	// 支配的かを 1 行で判別できるようにする。
	log.Printf("analyzer: callgraph breakdown: ssa=%s cha=%s bfs=%s convert=%s (pkgs=%d, cha nodes=%d, collected=%d, out nodes=%d edges=%d)",
		ssaTook, chaTook, bfsTook, time.Since(tConv),
		len(projectPkgs), len(cg.Nodes), len(collected), len(nodes), len(edges))

	return &domain.Graph{Nodes: nodes, Edges: edges}, nil
}

// walk は seeds から maxDepth ホップまで一方向に辿り、到達したノードを collected に加える。
// forward が true なら呼び出し先（Out）、false なら呼び出し元（In）の方向に辿る。
// プロジェクト外（loadedSet に無いパッケージ）の関数は辿らない。
func walk(seeds []*callgraph.Node, forward bool, maxDepth int, loadedSet map[string]struct{}, collected map[*callgraph.Node]struct{}) {
	type entry struct {
		node  *callgraph.Node
		depth int
	}

	visited := make(map[*callgraph.Node]struct{}, len(seeds))
	queue := make([]entry, 0, len(seeds))
	for _, s := range seeds {
		visited[s] = struct{}{}
		queue = append(queue, entry{s, 0})
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		collected[cur.node] = struct{}{}

		if cur.depth >= maxDepth {
			continue
		}

		edges := cur.node.In
		if forward {
			edges = cur.node.Out
		}
		for _, edge := range edges {
			next := edge.Caller
			if forward {
				next = edge.Callee
			}
			if !isInLoadedSet(next.Func, loadedSet) {
				continue
			}
			if _, ok := visited[next]; ok {
				continue
			}
			visited[next] = struct{}{}
			queue = append(queue, entry{next, cur.depth + 1})
		}
	}
}

func isInLoadedSet(fn *ssa.Function, loadedSet map[string]struct{}) bool {
	if fn == nil || fn.Package() == nil {
		return false
	}
	_, ok := loadedSet[fn.Package().Pkg.Path()]
	return ok
}

// rootFunc は無名関数（クロージャ）を囲うトップレベル関数まで遡って返す。
// SSA はクロージャを `親関数名$連番` で命名し独立した関数として扱うため、
// グラフ上で `regroupBy$1` のような別ノードに分裂する。これを親 1 ノードへ
// 畳み込むため、fn.Parent() が nil（=トップレベル）になるまで遡る。
// ネストしたクロージャも最上位の関数まで遡る。
func rootFunc(fn *ssa.Function) *ssa.Function {
	for fn.Parent() != nil {
		fn = fn.Parent()
	}
	return fn
}

// toNodeID は関数の一意な識別子を返す。
// fn.Name() はメソッドだとレシーバ型を含まない短い名前（例: "M"）になり、
// 同一パッケージ内で同名メソッドを持つ複数の型があると NodeID が衝突して
// 別々の関数が 1 ノードに統合されてしまう。クロージャ畳み込み（rootFunc）でも
// 同名メソッド内のクロージャが同じく衝突する。これを防ぐため、レシーバ型まで
// 含む fn.String()（= RelString(nil)、例: "(*pkg.T).M"）を識別子に使う。
func toNodeID(fn *ssa.Function) domain.NodeID {
	return domain.NodeID(fn.String())
}

func pkgPath(fn *ssa.Function) string {
	if fn.Package() == nil {
		return ""
	}
	return fn.Package().Pkg.Path()
}
