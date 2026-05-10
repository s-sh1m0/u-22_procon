package analyzer

import (
	"sort"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// MergeWithDiffStatus は base / head のグラフを合成し、各ノード/エッジに DiffStatus を付与する。
//
//   - base にも head にも存在 → DiffStatusExisting
//   - head にのみ存在        → DiffStatusAdded
//   - base にのみ存在        → DiffStatusRemoved
//
// 出力ノードのフィールド（File, Line など）は head にあれば head の値を優先し、無ければ base の値を採用する。
// removed ノードの Changed は false にリセットする（PR の changed ファイルは head 基準のため）。
//
// nil を渡された側は空のグラフとして扱う。
// 出力の Nodes/Edges はキー昇順でソートされており、テストの再現性を担保する。
func MergeWithDiffStatus(base, head *domain.Graph) *domain.Graph {
	if base == nil {
		base = &domain.Graph{}
	}
	if head == nil {
		head = &domain.Graph{}
	}

	nodeMap := make(map[domain.NodeID]domain.Node, len(base.Nodes)+len(head.Nodes))
	for _, n := range head.Nodes {
		nn := n
		nn.DiffStatus = domain.DiffStatusAdded // base にも存在すれば後で existing に降格
		nodeMap[n.ID] = nn
	}
	for _, n := range base.Nodes {
		if existing, ok := nodeMap[n.ID]; ok {
			existing.DiffStatus = domain.DiffStatusExisting
			nodeMap[n.ID] = existing
			continue
		}
		nn := n
		nn.Changed = false
		nn.DiffStatus = domain.DiffStatusRemoved
		nodeMap[n.ID] = nn
	}

	type edgeKey struct {
		from domain.NodeID
		to   domain.NodeID
	}
	edgeStatus := make(map[edgeKey]domain.DiffStatus, len(base.Edges)+len(head.Edges))
	for _, e := range head.Edges {
		edgeStatus[edgeKey{e.From, e.To}] = domain.DiffStatusAdded
	}
	for _, e := range base.Edges {
		k := edgeKey{e.From, e.To}
		if _, ok := edgeStatus[k]; ok {
			edgeStatus[k] = domain.DiffStatusExisting
		} else {
			edgeStatus[k] = domain.DiffStatusRemoved
		}
	}

	nodes := make([]domain.Node, 0, len(nodeMap))
	for _, n := range nodeMap {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })

	edges := make([]domain.Edge, 0, len(edgeStatus))
	for k, st := range edgeStatus {
		edges = append(edges, domain.Edge{From: k.from, To: k.to, Status: st})
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})

	return &domain.Graph{Nodes: nodes, Edges: edges}
}
