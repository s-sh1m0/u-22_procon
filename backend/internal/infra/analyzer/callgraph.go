package analyzer

import (
	"context"

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
// source_tree.go（issue #4）が返す PreparedSource.Packages と
// PreparedSource.ChangedPackages をそのまま渡せる設計にする。
type CallGraphBuilder interface {
	Build(ctx context.Context, pkgs []*packages.Package, changedPkgIDs []string) (*domain.Graph, error)
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
// stdlib・外部ライブラリはフィルタリングして含めない。
func (b *GoCallGraphBuilder) Build(_ context.Context, pkgs []*packages.Package, changedPkgIDs []string) (*domain.Graph, error) {
	if len(changedPkgIDs) == 0 {
		return &domain.Graph{}, nil
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
	nodeIDMap := make(map[*callgraph.Node]domain.NodeID, len(collected))
	var nodes []domain.Node
	for cgNode := range collected {
		fn := cgNode.Func
		nid := toNodeID(fn)
		nodeIDMap[cgNode] = nid

		changed := false
		if fn.Package() != nil {
			_, changed = changedSet[fn.Package().Pkg.Path()]
		}

		pos := prog.Fset.Position(fn.Pos())
		nodes = append(nodes, domain.Node{
			ID:      nid,
			Name:    fn.Name(),
			Package: pkgPath(fn),
			File:    pos.Filename,
			Line:    pos.Line,
			Changed: changed,
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
			key := [2]domain.NodeID{callerID, calleeID}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, domain.Edge{From: callerID, To: calleeID})
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
