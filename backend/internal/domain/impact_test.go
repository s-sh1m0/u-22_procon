package domain

import (
	"reflect"
	"testing"
)

func TestImpactSubgraph(t *testing.T) {
	node := func(id, file string, changed bool, st DiffStatus) Node {
		return Node{ID: NodeID(id), File: file, Changed: changed, DiffStatus: st}
	}
	edge := func(from, to string, st DiffStatus) Edge {
		return Edge{From: NodeID(from), To: NodeID(to), Status: st}
	}
	iface := func(id, file string) Node {
		return Node{ID: NodeID(id), Kind: NodeKindInterfaceMethod, File: file, DiffStatus: DiffStatusExisting}
	}
	impl := func(from, to string) Edge {
		return Edge{From: NodeID(from), To: NodeID(to), Kind: EdgeKindImplements, Status: DiffStatusExisting}
	}
	modified := []ChangedFile{{Filename: "a.go", Status: FileStatusModified}}

	tests := []struct {
		name       string
		g          *Graph
		cycles     []Cycle
		changed    []ChangedFile
		wantNodes  map[NodeID]DiffStatus
		wantEdges  map[[2]NodeID]DiffStatus
		wantCycles []int
	}{
		{
			name: "変更ノードとその直接の呼び出し元・呼び出し先だけを残す",
			g: &Graph{
				Nodes: []Node{
					node("Caller2", "c.go", false, DiffStatusExisting),
					node("Caller", "c.go", false, DiffStatusExisting),
					node("S", "a.go", true, DiffStatusExisting),
					node("Callee", "b.go", false, DiffStatusExisting),
					node("Callee2", "b.go", false, DiffStatusExisting),
					node("Unrelated", "a.go", false, DiffStatusExisting),
				},
				Edges: []Edge{
					edge("Caller2", "Caller", DiffStatusExisting),
					edge("Caller", "S", DiffStatusExisting),
					edge("S", "Callee", DiffStatusExisting),
					edge("Callee", "Callee2", DiffStatusExisting),
				},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"Caller": DiffStatusExisting,
				"S":      DiffStatusExisting,
				"Callee": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"Caller", "S"}: DiffStatusExisting,
				{"S", "Callee"}: DiffStatusExisting,
			},
			wantCycles: []int{},
		},
		{
			name: "変更ファイルの added / removed ノードも起点にする",
			g: &Graph{
				Nodes: []Node{
					node("A", "a.go", false, DiffStatusAdded),
					node("N", "b.go", false, DiffStatusExisting),
					node("R", "a.go", false, DiffStatusRemoved),
					node("M", "b.go", false, DiffStatusExisting),
				},
				Edges: []Edge{
					edge("A", "N", DiffStatusAdded),
					edge("M", "R", DiffStatusRemoved),
				},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"A": DiffStatusAdded,
				"N": DiffStatusExisting,
				"R": DiffStatusRemoved,
				"M": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"A", "N"}: DiffStatusAdded,
				{"M", "R"}: DiffStatusRemoved,
			},
			wantCycles: []int{},
		},
		{
			name: "rename 前のパスにある removed ノードも起点にする",
			g: &Graph{
				Nodes: []Node{node("R", "old.go", false, DiffStatusRemoved)},
			},
			changed:    []ChangedFile{{Filename: "new.go", PreviousFilename: "old.go", Status: FileStatusRenamed}},
			wantNodes:  map[NodeID]DiffStatus{"R": DiffStatusRemoved},
			wantEdges:  map[[2]NodeID]DiffStatus{},
			wantCycles: []int{},
		},
		{
			name: "変更ファイル外の added / removed は existing に戻し起点にしない",
			g: &Graph{
				Nodes: []Node{
					node("S", "a.go", true, DiffStatusExisting),
					node("X", "x.go", false, DiffStatusAdded),
					node("Y", "y.go", false, DiffStatusExisting),
					node("Z", "z.go", false, DiffStatusRemoved),
				},
				Edges: []Edge{
					edge("S", "X", DiffStatusAdded),
					edge("X", "Y", DiffStatusAdded),
				},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"S": DiffStatusExisting,
				"X": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"S", "X"}: DiffStatusAdded,
			},
			wantCycles: []int{},
		},
		{
			name: "両端とも変更ファイル外のエッジの added は existing に戻す",
			g: &Graph{
				Nodes: []Node{
					node("S", "a.go", true, DiffStatusExisting),
					node("N1", "b.go", false, DiffStatusExisting),
					node("N2", "b.go", false, DiffStatusExisting),
				},
				Edges: []Edge{
					edge("S", "N1", DiffStatusAdded),
					edge("S", "N2", DiffStatusExisting),
					edge("N1", "N2", DiffStatusAdded),
				},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"S":  DiffStatusExisting,
				"N1": DiffStatusExisting,
				"N2": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"S", "N1"}:  DiffStatusAdded,
				{"S", "N2"}:  DiffStatusExisting,
				{"N1", "N2"}: DiffStatusExisting,
			},
			wantCycles: []int{},
		},
		{
			name: "起点を含む新規循環は構成ノードも残し、残らない循環は返さない",
			g: &Graph{
				Nodes: []Node{
					node("S", "a.go", true, DiffStatusExisting),
					node("A", "b.go", false, DiffStatusExisting),
					node("B", "b.go", false, DiffStatusExisting),
					node("C", "b.go", false, DiffStatusExisting),
					node("X", "x.go", false, DiffStatusExisting),
					node("Y", "x.go", false, DiffStatusExisting),
				},
				Edges: []Edge{
					edge("S", "A", DiffStatusAdded),
					edge("A", "B", DiffStatusExisting),
					edge("B", "C", DiffStatusExisting),
					edge("C", "S", DiffStatusExisting),
					edge("X", "Y", DiffStatusExisting),
					edge("Y", "X", DiffStatusExisting),
				},
			},
			cycles: []Cycle{
				{ID: 0, Nodes: []NodeID{"A", "B", "C", "S"}, IsNew: true},
				{ID: 1, Nodes: []NodeID{"X", "Y"}, IsNew: false},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"S": DiffStatusExisting,
				"A": DiffStatusExisting,
				"B": DiffStatusExisting,
				"C": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"S", "A"}: DiffStatusAdded,
				{"A", "B"}: DiffStatusExisting,
				{"B", "C"}: DiffStatusExisting,
				{"C", "S"}: DiffStatusExisting,
			},
			wantCycles: []int{0},
		},
		{
			name: "起点が呼ぶ interface メソッドの実装を残し、別の呼び出し元は残さない",
			g: &Graph{
				Nodes: []Node{
					node("S", "a.go", true, DiffStatusExisting),
					iface("I", "i.go"),
					node("Impl1", "b.go", false, DiffStatusExisting),
					node("Impl2", "b.go", false, DiffStatusExisting),
					node("OtherCaller", "c.go", false, DiffStatusExisting),
					node("ImplCallee", "b.go", false, DiffStatusExisting),
				},
				Edges: []Edge{
					edge("S", "I", DiffStatusAdded),
					impl("I", "Impl1"),
					impl("I", "Impl2"),
					edge("OtherCaller", "I", DiffStatusExisting),
					edge("Impl1", "ImplCallee", DiffStatusExisting),
				},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"S":     DiffStatusExisting,
				"I":     DiffStatusExisting,
				"Impl1": DiffStatusExisting,
				"Impl2": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"S", "I"}:     DiffStatusAdded,
				{"I", "Impl1"}: DiffStatusExisting,
				{"I", "Impl2"}: DiffStatusExisting,
			},
			wantCycles: []int{},
		},
		{
			name: "起点を実装に持つ interface メソッドの呼び出し元を残し、別の実装は残さない",
			g: &Graph{
				Nodes: []Node{
					node("S", "a.go", true, DiffStatusExisting),
					iface("I", "i.go"),
					node("Caller", "c.go", false, DiffStatusExisting),
					node("OtherImpl", "b.go", false, DiffStatusExisting),
				},
				Edges: []Edge{
					edge("Caller", "I", DiffStatusExisting),
					impl("I", "S"),
					impl("I", "OtherImpl"),
				},
			},
			changed: modified,
			wantNodes: map[NodeID]DiffStatus{
				"S":      DiffStatusExisting,
				"I":      DiffStatusExisting,
				"Caller": DiffStatusExisting,
			},
			wantEdges: map[[2]NodeID]DiffStatus{
				{"Caller", "I"}: DiffStatusExisting,
				{"I", "S"}:      DiffStatusExisting,
			},
			wantCycles: []int{},
		},
		{
			name:       "nil グラフは空グラフを返す",
			g:          nil,
			changed:    modified,
			wantNodes:  map[NodeID]DiffStatus{},
			wantEdges:  map[[2]NodeID]DiffStatus{},
			wantCycles: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, gotCycles := ImpactSubgraph(tc.g, tc.cycles, tc.changed)

			gotNodes := make(map[NodeID]DiffStatus, len(got.Nodes))
			for _, n := range got.Nodes {
				gotNodes[n.ID] = n.DiffStatus
			}
			if !reflect.DeepEqual(gotNodes, tc.wantNodes) {
				t.Errorf("nodes = %v, want %v", gotNodes, tc.wantNodes)
			}

			gotEdges := make(map[[2]NodeID]DiffStatus, len(got.Edges))
			for _, e := range got.Edges {
				gotEdges[[2]NodeID{e.From, e.To}] = e.Status
			}
			if !reflect.DeepEqual(gotEdges, tc.wantEdges) {
				t.Errorf("edges = %v, want %v", gotEdges, tc.wantEdges)
			}

			var gotCycleIDs []int
			if gotCycles != nil {
				gotCycleIDs = make([]int, 0, len(gotCycles))
			}
			for _, c := range gotCycles {
				gotCycleIDs = append(gotCycleIDs, c.ID)
			}
			if !reflect.DeepEqual(gotCycleIDs, tc.wantCycles) {
				t.Errorf("cycles = %v, want %v", gotCycleIDs, tc.wantCycles)
			}
		})
	}
}

func TestImpactSubgraph_DoesNotMutateInput(t *testing.T) {
	g := &Graph{
		Nodes: []Node{
			{ID: "S", File: "a.go", Changed: true, DiffStatus: DiffStatusExisting},
			{ID: "N1", File: "b.go", DiffStatus: DiffStatusExisting},
			{ID: "N2", File: "b.go", DiffStatus: DiffStatusAdded},
		},
		Edges: []Edge{
			{From: "S", To: "N1", Status: DiffStatusExisting},
			{From: "S", To: "N2", Status: DiffStatusExisting},
			{From: "N1", To: "N2", Status: DiffStatusAdded},
		},
	}
	ImpactSubgraph(g, nil, []ChangedFile{{Filename: "a.go"}})

	if g.Nodes[2].DiffStatus != DiffStatusAdded {
		t.Errorf("input node N2 DiffStatus mutated to %q", g.Nodes[2].DiffStatus)
	}
	if g.Edges[2].Status != DiffStatusAdded {
		t.Errorf("input edge N1->N2 Status mutated to %q", g.Edges[2].Status)
	}
}
