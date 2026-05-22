package usecase

import (
	"sort"
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

func sampleGraph() domain.Graph {
	return domain.Graph{
		Nodes: []domain.Node{
			{ID: "a", Name: "A", Package: "example.com/pkg/foo", File: "foo/a.go"},
			{ID: "b", Name: "B", Package: "example.com/pkg/foo", File: "foo/a.go"},
			{ID: "c", Name: "C", Package: "example.com/pkg/foo", File: "foo/c.go"},
			{ID: "d", Name: "D", Package: "example.com/pkg/bar", File: "bar/d.go"},
			{ID: "e", Name: "E", Package: "", File: ""},
		},
	}
}

func nodeIDs(c domain.Cluster) []string {
	out := make([]string, len(c.Nodes))
	for i, n := range c.Nodes {
		out[i] = string(n)
	}
	sort.Strings(out)
	return out
}

func TestRegroupBy_Package(t *testing.T) {
	got := regroupBy(sampleGraph(), domain.ClusterModePackage)
	if len(got) != 3 {
		t.Fatalf("want 3 clusters, got %d (%+v)", len(got), got)
	}
	// サイズ降順: foo(3) > bar(1) = ""(1), tie-break で空文字 < "example.com/pkg/bar"
	if got[0].Label != "foo" || len(got[0].Nodes) != 3 {
		t.Errorf("cluster[0] = %+v; want foo/3", got[0])
	}
	if got[1].Label != "(no package)" || len(got[1].Nodes) != 1 {
		t.Errorf("cluster[1] = %+v; want (no package)/1", got[1])
	}
	if got[2].Label != "bar" {
		t.Errorf("cluster[2].Label = %q; want bar", got[2].Label)
	}
	foo := nodeIDs(got[0])
	if foo[0] != "a" || foo[1] != "b" || foo[2] != "c" {
		t.Errorf("foo cluster nodes = %v; want [a b c]", foo)
	}
}

func TestRegroupBy_File(t *testing.T) {
	got := regroupBy(sampleGraph(), domain.ClusterModeFile)
	if len(got) != 4 {
		t.Fatalf("want 4 clusters, got %d", len(got))
	}
	// foo/a.go(2) が先頭
	if got[0].Label != "a.go" || len(got[0].Nodes) != 2 {
		t.Errorf("cluster[0] = %+v; want a.go/2", got[0])
	}
	// tie-break: 空文字 < bar/d.go < foo/c.go
	if got[1].Label != "(no file)" {
		t.Errorf("cluster[1].Label = %q; want (no file)", got[1].Label)
	}
	if got[2].Label != "d.go" {
		t.Errorf("cluster[2].Label = %q; want d.go", got[2].Label)
	}
	if got[3].Label != "c.go" {
		t.Errorf("cluster[3].Label = %q; want c.go", got[3].Label)
	}
}

func TestApplyClusterMode_LouvainLabelsClusters(t *testing.T) {
	orig := &domain.ClusterResult{
		Clusters: []domain.Cluster{{ID: 0, Label: "raw", Nodes: []domain.NodeID{"a", "b"}}},
		Graph:    sampleGraph(),
		Cycles:   []domain.Cycle{{ID: 3, Nodes: []domain.NodeID{"a", "b"}, IsNew: true}},
	}
	got := applyClusterMode(orig, domain.ClusterModeLouvain)
	if got == nil {
		t.Fatal("Louvain mode returned nil")
	}
	// Louvain モードはクラスタ構成（ノードの所属）を変えずにラベルだけ付け直す。
	if len(got.Clusters) != 1 || len(got.Clusters[0].Nodes) != 2 {
		t.Errorf("clusters should be preserved, got %+v", got.Clusters)
	}
	if got.Clusters[0].Label == "raw" {
		t.Error("Louvain mode should relabel clusters via labelClusters, but kept the input label")
	}
	// Graph と Cycles はそのまま保持される。
	if len(got.Graph.Nodes) != len(orig.Graph.Nodes) {
		t.Errorf("Graph not preserved: got %d nodes, want %d", len(got.Graph.Nodes), len(orig.Graph.Nodes))
	}
	if len(got.Cycles) != 1 || got.Cycles[0].ID != 3 {
		t.Errorf("Cycles not preserved: %+v", got.Cycles)
	}
}

func TestApplyClusterMode_NilResult(t *testing.T) {
	if got := applyClusterMode(nil, domain.ClusterModePackage); got != nil {
		t.Errorf("nil result should pass through, got %+v", got)
	}
}

func TestApplyClusterMode_PreservesCycles(t *testing.T) {
	orig := &domain.ClusterResult{
		Clusters: []domain.Cluster{{ID: 0, Label: "louvain-cluster", Nodes: []domain.NodeID{"a", "b"}}},
		Graph:    sampleGraph(),
		Cycles:   []domain.Cycle{{ID: 7, Nodes: []domain.NodeID{"a", "b"}, IsNew: true}},
	}
	got := applyClusterMode(orig, domain.ClusterModePackage)
	if len(got.Cycles) != 1 || got.Cycles[0].ID != 7 {
		t.Errorf("Cycles not preserved: %+v", got.Cycles)
	}
	if got.Clusters[0].Label == "louvain-cluster" {
		t.Error("Clusters should be regrouped, but kept Louvain label")
	}
}
