package analyzer

import (
	"sort"
	"strings"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// DetectCycles はコールグラフ上の循環参照（強連結成分）を検出する。
// size >= 2 の SCC のみを循環として返す（自己ループは除外）。
// 各 NodeID リストは ID 昇順でソート済み、テストの再現性を担保する。
func DetectCycles(g *domain.Graph) [][]domain.NodeID {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}

	dg := simple.NewDirectedGraph()
	idToInt := make(map[domain.NodeID]int64, len(g.Nodes))
	intToID := make(map[int64]domain.NodeID, len(g.Nodes))
	for i, n := range g.Nodes {
		id := int64(i)
		dg.AddNode(simple.Node(id))
		idToInt[n.ID] = id
		intToID[id] = n.ID
	}
	for _, e := range g.Edges {
		u, ok1 := idToInt[e.From]
		v, ok2 := idToInt[e.To]
		if !ok1 || !ok2 || u == v {
			continue
		}
		// 同一 (u,v) の重複追加を避ける
		if dg.HasEdgeFromTo(u, v) {
			continue
		}
		dg.SetEdge(simple.Edge{F: simple.Node(u), T: simple.Node(v)})
	}

	sccs := topo.TarjanSCC(dg)
	out := make([][]domain.NodeID, 0)
	for _, scc := range sccs {
		if len(scc) < 2 {
			continue
		}
		ids := make([]domain.NodeID, 0, len(scc))
		for _, n := range scc {
			ids = append(ids, intToID[n.ID()])
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		out = append(out, ids)
	}

	// SCC 自体も先頭 NodeID 昇順でソート（テストの再現性）。
	sort.Slice(out, func(i, j int) bool { return out[i][0] < out[j][0] })
	return out
}

// DetectNewCycles は base/head のグラフを比較し、Cycle のリストを返す。
// 同一性判定はノード ID 集合の完全一致で行う。Base に同じ集合が無い Head の
// cycle を IsNew=true、両方にある cycle を IsNew=false で返す。
//
// Issue #27 のアラート対象は IsNew=true のみ。
func DetectNewCycles(base, head *domain.Graph) []domain.Cycle {
	headCycles := DetectCycles(head)
	baseCycles := DetectCycles(base)

	baseSet := make(map[string]struct{}, len(baseCycles))
	for _, c := range baseCycles {
		baseSet[cycleKey(c)] = struct{}{}
	}

	out := make([]domain.Cycle, 0, len(headCycles))
	for i, c := range headCycles {
		_, existed := baseSet[cycleKey(c)]
		out = append(out, domain.Cycle{
			ID:    i,
			Nodes: c,
			IsNew: !existed,
		})
	}
	return out
}

// cycleKey は ID ソート済みの NodeID リストから一意なキーを作る。
func cycleKey(ids []domain.NodeID) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = string(id)
	}
	return strings.Join(parts, "\x00")
}
