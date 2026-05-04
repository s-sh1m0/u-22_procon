package cluster

import (
	"gonum.org/v1/gonum/graph/simple"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// gonumAdapter は domain.Graph と gonum WeightedUndirectedGraph の双方向 ID マッピングを保持する。
type gonumAdapter struct {
	g        *simple.WeightedUndirectedGraph
	nodeToID map[domain.NodeID]int64
	idToNode map[int64]domain.NodeID
}

// newGonumAdapter は domain.Graph を重み付き無向グラフに変換する。
// 自己ループは除外し、双方向辺（A→B かつ B→A）は重み 1.0 の単一無向辺に統合する。
// エッジを持たない孤立ノードも gonum 側に保持する。
func newGonumAdapter(dg *domain.Graph) *gonumAdapter {
	g := simple.NewWeightedUndirectedGraph(0, 0)
	nodeToID := make(map[domain.NodeID]int64, len(dg.Nodes))
	idToNode := make(map[int64]domain.NodeID, len(dg.Nodes))

	for i, n := range dg.Nodes {
		id := int64(i)
		nodeToID[n.ID] = id
		idToNode[id] = n.ID
		g.AddNode(simple.Node(id))
	}

	// 重複辺除去用セット（無向なので min/max でキー化）
	seen := make(map[[2]int64]struct{})
	for _, e := range dg.Edges {
		u, ok1 := nodeToID[e.From]
		v, ok2 := nodeToID[e.To]
		if !ok1 || !ok2 {
			continue
		}
		// 自己ループを除外
		if u == v {
			continue
		}
		lo, hi := u, v
		if lo > hi {
			lo, hi = hi, lo
		}
		key := [2]int64{lo, hi}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		g.SetWeightedEdge(g.NewWeightedEdge(simple.Node(u), simple.Node(v), 1.0))
	}

	return &gonumAdapter{g: g, nodeToID: nodeToID, idToNode: idToNode}
}
