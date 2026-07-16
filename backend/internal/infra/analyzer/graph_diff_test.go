package analyzer

import (
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func node(id, name string) domain.Node {
	return domain.Node{ID: domain.NodeID(id), Name: name, DiffStatus: domain.DiffStatusExisting}
}

func edge(from, to string) domain.Edge {
	return domain.Edge{From: domain.NodeID(from), To: domain.NodeID(to), Status: domain.DiffStatusExisting}
}

func TestMergeWithDiffStatus_NodesUnion(t *testing.T) {
	base := &domain.Graph{
		Nodes: []domain.Node{node("a", "A"), node("b", "B")},
	}
	head := &domain.Graph{
		Nodes: []domain.Node{node("b", "B"), node("c", "C")},
	}

	got := MergeWithDiffStatus(base, head)

	want := map[domain.NodeID]domain.DiffStatus{
		"a": domain.DiffStatusRemoved,
		"b": domain.DiffStatusExisting,
		"c": domain.DiffStatusAdded,
	}
	if len(got.Nodes) != len(want) {
		t.Fatalf("len(nodes)=%d want %d", len(got.Nodes), len(want))
	}
	for _, n := range got.Nodes {
		if n.DiffStatus != want[n.ID] {
			t.Errorf("node %s: got %s want %s", n.ID, n.DiffStatus, want[n.ID])
		}
	}
}

func TestMergeWithDiffStatus_EdgesUnion(t *testing.T) {
	base := &domain.Graph{
		Edges: []domain.Edge{edge("a", "b"), edge("b", "c")},
	}
	head := &domain.Graph{
		Edges: []domain.Edge{edge("b", "c"), edge("c", "d")},
	}

	got := MergeWithDiffStatus(base, head)

	want := map[[2]domain.NodeID]domain.DiffStatus{
		{"a", "b"}: domain.DiffStatusRemoved,
		{"b", "c"}: domain.DiffStatusExisting,
		{"c", "d"}: domain.DiffStatusAdded,
	}
	if len(got.Edges) != len(want) {
		t.Fatalf("len(edges)=%d want %d", len(got.Edges), len(want))
	}
	for _, e := range got.Edges {
		k := [2]domain.NodeID{e.From, e.To}
		if e.Status != want[k] {
			t.Errorf("edge %s->%s: got %s want %s", e.From, e.To, e.Status, want[k])
		}
	}
}

func TestMergeWithDiffStatus_RemovedNodeChangedReset(t *testing.T) {
	base := &domain.Graph{
		Nodes: []domain.Node{{ID: "a", Changed: true, DiffStatus: domain.DiffStatusExisting}},
	}
	head := &domain.Graph{}

	got := MergeWithDiffStatus(base, head)

	if len(got.Nodes) != 1 {
		t.Fatalf("len(nodes)=%d want 1", len(got.Nodes))
	}
	if got.Nodes[0].Changed {
		t.Error("removed node should have Changed=false")
	}
	if got.Nodes[0].DiffStatus != domain.DiffStatusRemoved {
		t.Errorf("removed node DiffStatus=%s want removed", got.Nodes[0].DiffStatus)
	}
}

func TestMergeWithDiffStatus_HeadFieldsPreferred(t *testing.T) {
	base := &domain.Graph{
		Nodes: []domain.Node{{ID: "a", File: "old.go", Line: 1}},
	}
	head := &domain.Graph{
		Nodes: []domain.Node{{ID: "a", File: "new.go", Line: 10, Changed: true}},
	}

	got := MergeWithDiffStatus(base, head)

	if len(got.Nodes) != 1 {
		t.Fatalf("len(nodes)=%d want 1", len(got.Nodes))
	}
	n := got.Nodes[0]
	if n.File != "new.go" || n.Line != 10 {
		t.Errorf("head fields not preferred: %+v", n)
	}
	if !n.Changed {
		t.Error("Changed should be preserved from head")
	}
	if n.DiffStatus != domain.DiffStatusExisting {
		t.Errorf("DiffStatus=%s want existing", n.DiffStatus)
	}
}

func TestMergeWithDiffStatus_NilInputs(t *testing.T) {
	got := MergeWithDiffStatus(nil, nil)
	if got == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(got.Nodes) != 0 || len(got.Edges) != 0 {
		t.Errorf("expected empty graph, got %+v", got)
	}
}

func TestMergeWithDiffStatus_DeterministicOrder(t *testing.T) {
	// 入力順を変えても出力が同じになることを確認
	g1 := MergeWithDiffStatus(
		&domain.Graph{Nodes: []domain.Node{node("c", "C"), node("a", "A")}},
		&domain.Graph{Nodes: []domain.Node{node("b", "B"), node("a", "A")}},
	)
	g2 := MergeWithDiffStatus(
		&domain.Graph{Nodes: []domain.Node{node("a", "A"), node("c", "C")}},
		&domain.Graph{Nodes: []domain.Node{node("a", "A"), node("b", "B")}},
	)
	if len(g1.Nodes) != len(g2.Nodes) {
		t.Fatalf("len mismatch %d vs %d", len(g1.Nodes), len(g2.Nodes))
	}
	for i := range g1.Nodes {
		if g1.Nodes[i].ID != g2.Nodes[i].ID {
			t.Errorf("idx %d: %s vs %s", i, g1.Nodes[i].ID, g2.Nodes[i].ID)
		}
	}
}
