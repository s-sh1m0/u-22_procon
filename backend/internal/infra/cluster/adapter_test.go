package cluster

import (
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func newTestGraph(nodes []domain.Node, pairs [][2]domain.NodeID) *domain.Graph {
	edges := make([]domain.Edge, 0, len(pairs))
	for _, p := range pairs {
		edges = append(edges, domain.Edge{From: p[0], To: p[1]})
	}
	return &domain.Graph{Nodes: nodes, Edges: edges}
}

func TestNewGonumAdapter_NodeIDMapping(t *testing.T) {
	nodes := []domain.Node{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
		{ID: "c", Name: "C"},
	}
	g := newTestGraph(nodes, nil)
	a := newGonumAdapter(g)

	if len(a.nodeToID) != 3 {
		t.Fatalf("nodeToID len=%d, want 3", len(a.nodeToID))
	}
	// 双方向 round-trip
	for _, n := range nodes {
		id, ok := a.nodeToID[n.ID]
		if !ok {
			t.Errorf("nodeToID missing %q", n.ID)
			continue
		}
		got, ok2 := a.idToNode[id]
		if !ok2 || got != n.ID {
			t.Errorf("idToNode[%d]=%q, want %q", id, got, n.ID)
		}
	}
}

func TestNewGonumAdapter_DropsSelfLoops(t *testing.T) {
	nodes := []domain.Node{{ID: "a"}}
	g := newTestGraph(nodes, [][2]domain.NodeID{{"a", "a"}})
	a := newGonumAdapter(g)

	if a.g.Edges().Len() != 0 {
		t.Errorf("want 0 edges (self-loop dropped), got %d", a.g.Edges().Len())
	}
}

func TestNewGonumAdapter_DedupsBidirectionalEdges(t *testing.T) {
	nodes := []domain.Node{{ID: "a"}, {ID: "b"}}
	// A→B と B→A の両方を渡す
	g := newTestGraph(nodes, [][2]domain.NodeID{{"a", "b"}, {"b", "a"}})
	a := newGonumAdapter(g)

	if got := a.g.Edges().Len(); got != 1 {
		t.Errorf("want 1 edge (deduped), got %d", got)
	}
}

func TestNewGonumAdapter_PreservesIsolatedNodes(t *testing.T) {
	nodes := []domain.Node{{ID: "a"}, {ID: "isolated"}}
	g := newTestGraph(nodes, [][2]domain.NodeID{{"a", "a"}}) // 自己ループのみ
	a := newGonumAdapter(g)

	if got := a.g.Nodes().Len(); got != 2 {
		t.Errorf("want 2 nodes (isolated preserved), got %d", got)
	}
}
