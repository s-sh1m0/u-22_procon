package analyzer

import (
	"context"
	"path/filepath"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// defaultMaxDepth は変更関数から辿るホップ数のデフォルト上限。
const defaultMaxDepth = 3

// CallGraphBuilder はロード済みパッケージから呼び出しグラフを構築する抽象。
// source_tree.go が返す PreparedSource.Packages / ChangedPackages / ChangedFileAbsPaths をそのまま渡せる設計にする。
type CallGraphBuilder interface {
	Build(ctx context.Context, pkgs []*packages.Package, changedPkgIDs []string, changedFileAbsPaths []string, rootDir string) (*domain.Graph, error)
}

// GoCallGraphBuilder は golang.org/x/tools/go/callgraph を使う本番実装。
type GoCallGraphBuilder struct {
	// MaxDepth は変更関数から辿るホップ数の上限。0 以下でデフォルト値を使う。
	MaxDepth int
}

// NewGoCallGraphBuilder は GoCallGraphBuilder を返す。
func NewGoCallGraphBuilder() *GoCallGraphBuilder {
	return &GoCallGraphBuilder{MaxDepth: defaultMaxDepth}
}

// Build は pkgs から SSA を構築し、changedPkgIDs を起点に caller/callee グラフを返す。
// changedPkgIDs が空の場合は空の Graph を返す。
// changedFileAbsPaths は changed フラグをファイル単位で設定するために使う（パッケージ単位ではない）。
// stdlib・外部ライブラリはフィルタリングして含めない。
// rootDir はリポジトリのルートディレクトリの絶対パスで、ノードのファイルパスを相対パスに変換するために使用される。
func (b *GoCallGraphBuilder) Build(_ context.Context, pkgs []*packages.Package, changedPkgIDs []string, changedFileAbsPaths []string, rootDir string) (*domain.Graph, error) {
	if len(changedPkgIDs) == 0 {
		return &domain.Graph{}, nil
	}

	changedFileSet := make(map[string]struct{}, len(changedFileAbsPaths))
	for _, f := range changedFileAbsPaths {
		changedFileSet[f] = struct{}{}
	}

	maxDepth := b.MaxDepth
	if maxDepth <= 0 {
		maxDepth = defaultMaxDepth
	}

	// 1. プロジェクト内パッケージのみを抽出する。
	// NeedDeps によって pkgs には外部ライブラリも含まれるが、SSA 構築と
	// loadedSet はプロジェクト内パッケージのみを対象にする。
	// pkg.Module.Main が true のパッケージが自リポジトリのパッケージ。
	projectPkgs := make([]*packages.Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.Module != nil && pkg.Module.Main {
			projectPkgs = append(projectPkgs, pkg)
		}
	}

	// 2. SSA プログラムを構築（プロジェクト内パッケージのみ）
	prog, _ := ssautil.AllPackages(projectPkgs, ssa.InstantiateGenerics)
	prog.Build()

	// 3. CHA でコールグラフ構築（エントリポイント不要 → ライブラリにも適用可）
	cg := cha.CallGraph(prog)
	cg.DeleteSyntheticNodes()

	// 4. 変更パッケージ ID セット
	changedSet := make(map[string]struct{}, len(changedPkgIDs))
	for _, id := range changedPkgIDs {
		changedSet[id] = struct{}{}
	}

	// 5. プロジェクト内パッケージ ID セット（外部ライブラリ除外用）
	loadedSet := make(map[string]struct{}, len(projectPkgs))
	for _, pkg := range projectPkgs {
		loadedSet[pkg.ID] = struct{}{}
	}

	// 5. 変更パッケージの関数を起点に BFS（双方向、maxDepth ホップ）
	type entry struct {
		node  *callgraph.Node
		depth int
	}

	visited := make(map[*callgraph.Node]struct{})
	queue := []entry{}

	for fn, node := range cg.Nodes {
		if fn == nil || fn.Package() == nil {
			continue
		}
		if _, ok := changedSet[fn.Package().Pkg.Path()]; ok {
			if _, ok2 := visited[node]; !ok2 {
				visited[node] = struct{}{}
				queue = append(queue, entry{node, 0})
			}
		}
	}

	collected := make(map[*callgraph.Node]struct{})
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		collected[cur.node] = struct{}{}

		if cur.depth >= maxDepth {
			continue
		}

		// callee 方向
		for _, edge := range cur.node.Out {
			callee := edge.Callee
			if !isInLoadedSet(callee.Func, loadedSet) {
				continue
			}
			if _, ok := visited[callee]; !ok {
				visited[callee] = struct{}{}
				queue = append(queue, entry{callee, cur.depth + 1})
			}
		}
		// caller 方向
		for _, edge := range cur.node.In {
			caller := edge.Caller
			if !isInLoadedSet(caller.Func, loadedSet) {
				continue
			}
			if _, ok := visited[caller]; !ok {
				visited[caller] = struct{}{}
				queue = append(queue, entry{caller, cur.depth + 1})
			}
		}
	}

	// 6. 収集ノード・エッジを domain.Graph に変換
	// 無名関数（クロージャ）は囲うトップレベル親関数に畳み込む。複数の callgraph
	// ノード（親本体 + 各クロージャ）が同一 NodeID を指すため、ノードは重複排除し、
	// エッジは親 NodeID へ付け替えたうえで自己ループ・重複を除去する。
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
		_, fileChanged := changedFileSet[pos.Filename]
		relPath, err := filepath.Rel(rootDir, pos.Filename) // リポジトリからの相対パスにする
		if err != nil {
			relPath = pos.Filename
		}
		nodes = append(nodes, domain.Node{
			ID:         nid,
			Name:       fn.Name(),
			Package:    pkgPath(fn),
			File:       relPath,
			Line:       pos.Line,
			Changed:    fileChanged,
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

	return &domain.Graph{Nodes: nodes, Edges: edges}, nil
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

func toNodeID(fn *ssa.Function) domain.NodeID {
	if fn.Package() == nil {
		return domain.NodeID(fn.String())
	}
	return domain.NodeID(fn.Package().Pkg.Path() + "." + fn.Name())
}

func pkgPath(fn *ssa.Function) string {
	if fn.Package() == nil {
		return ""
	}
	return fn.Package().Pkg.Path()
}
