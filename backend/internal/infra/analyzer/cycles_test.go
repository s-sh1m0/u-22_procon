package analyzer

import (
	"reflect"
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func makeGraph(nodes []string, edges [][2]string) *domain.Graph {
	g := &domain.Graph{}
	for _, n := range nodes {
		g.Nodes = append(g.Nodes, domain.Node{ID: domain.NodeID(n)})
	}
	for _, e := range edges {
		g.Edges = append(g.Edges, domain.Edge{From: domain.NodeID(e[0]), To: domain.NodeID(e[1])})
	}
	return g
}

func TestDetectCycles(t *testing.T) {
	cases := []struct {
		name  string
		nodes []string
		edges [][2]string
		want  [][]domain.NodeID
	}{
		{
			name:  "no cycle",
			nodes: []string{"a", "b", "c"},
			edges: [][2]string{{"a", "b"}, {"b", "c"}},
			want:  [][]domain.NodeID{},
		},
		{
			name:  "two-node cycle",
			nodes: []string{"a", "b"},
			edges: [][2]string{{"a", "b"}, {"b", "a"}},
			want:  [][]domain.NodeID{{"a", "b"}},
		},
		{
			name:  "three-node cycle",
			nodes: []string{"a", "b", "c"},
			edges: [][2]string{{"a", "b"}, {"b", "c"}, {"c", "a"}},
			want:  [][]domain.NodeID{{"a", "b", "c"}},
		},
		{
			name:  "self loop is ignored",
			nodes: []string{"a"},
			edges: [][2]string{{"a", "a"}},
			want:  [][]domain.NodeID{},
		},
		{
			name:  "two disjoint cycles",
			nodes: []string{"a", "b", "c", "d"},
			edges: [][2]string{{"a", "b"}, {"b", "a"}, {"c", "d"}, {"d", "c"}},
			want:  [][]domain.NodeID{{"a", "b"}, {"c", "d"}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectCycles(makeGraph(tc.nodes, tc.edges))
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestDetectCycles_NilGraph(t *testing.T) {
	if got := DetectCycles(nil); got != nil {
		t.Errorf("nil graph: got %v want nil", got)
	}
}

func TestDetectNewCycles_NewlyIntroduced(t *testing.T) {
	base := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}})
	head := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}, {"b", "a"}})

	got := DetectNewCycles(base, head)
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if !got[0].IsNew {
		t.Error("expected IsNew=true")
	}
	if !reflect.DeepEqual(got[0].Nodes, []domain.NodeID{"a", "b"}) {
		t.Errorf("nodes=%v", got[0].Nodes)
	}
}

func TestDetectNewCycles_PreExisting(t *testing.T) {
	base := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}, {"b", "a"}})
	head := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}, {"b", "a"}})

	got := DetectNewCycles(base, head)
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if got[0].IsNew {
		t.Error("expected IsNew=false")
	}
}

func TestDetectNewCycles_NodeAddedToCycle(t *testing.T) {
	// Base: a->b->a (cycle {a,b})
	// Head: a->b->c->a (cycle {a,b,c}) — 集合が異なるので新規 cycle
	base := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}, {"b", "a"}})
	head := makeGraph([]string{"a", "b", "c"}, [][2]string{{"a", "b"}, {"b", "c"}, {"c", "a"}})

	got := DetectNewCycles(base, head)
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if !got[0].IsNew {
		t.Error("expected IsNew=true (different node set)")
	}
}

func TestDetectNewCycles_CycleRemoved(t *testing.T) {
	// Base に cycle、Head に無い → 出力には含めない（issue 要件は head 側 cycle の列挙）
	base := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}, {"b", "a"}})
	head := makeGraph([]string{"a", "b"}, [][2]string{{"a", "b"}})

	got := DetectNewCycles(base, head)
	if len(got) != 0 {
		t.Errorf("expected empty, got %+v", got)
	}
}
