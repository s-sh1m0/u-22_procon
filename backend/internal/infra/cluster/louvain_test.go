package cluster

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// buildK3 は 3 ノードの完全グラフを構築するヘルパー。
func buildK3(prefix string) ([]domain.Node, [][2]domain.NodeID) {
	ids := []domain.NodeID{
		domain.NodeID(prefix + "1"),
		domain.NodeID(prefix + "2"),
		domain.NodeID(prefix + "3"),
	}
	nodes := []domain.Node{
		{ID: ids[0], Name: prefix + "1"},
		{ID: ids[1], Name: prefix + "2"},
		{ID: ids[2], Name: prefix + "3"},
	}
	pairs := [][2]domain.NodeID{
		{ids[0], ids[1]}, {ids[1], ids[0]},
		{ids[1], ids[2]}, {ids[2], ids[1]},
		{ids[0], ids[2]}, {ids[2], ids[0]},
	}
	return nodes, pairs
}

func TestLouvainClusterer_TwoCliques(t *testing.T) {
	nodesA, pairsA := buildK3("A")
	nodesB, pairsB := buildK3("B")

	// 2 クリーク間を bridge エッジ 1 本で繋ぐ
	allNodes := append(nodesA, nodesB...)
	allPairs := append(pairsA, pairsB...)
	allPairs = append(allPairs, [2]domain.NodeID{"A1", "B1"})

	g := newTestGraph(allNodes, allPairs)

	c := NewLouvainClusterer()
	result, err := c.Cluster(context.Background(), g)
	if err != nil {
		t.Fatalf("Cluster: %v", err)
	}

	if len(result.Clusters) < 2 {
		t.Errorf("want >= 2 clusters for two cliques, got %d", len(result.Clusters))
	}
}

func TestLouvainClusterer_EmptyGraph(t *testing.T) {
	cases := []*domain.Graph{
		nil,
		{},
	}
	c := NewLouvainClusterer()
	for _, g := range cases {
		result, err := c.Cluster(context.Background(), g)
		if err != nil {
			t.Fatalf("Cluster: %v", err)
		}
		if len(result.Clusters) != 0 {
			t.Errorf("want 0 clusters for empty graph, got %d", len(result.Clusters))
		}
	}
}

func TestLouvainClusterer_Determinism(t *testing.T) {
	nodesA, pairsA := buildK3("A")
	nodesB, pairsB := buildK3("B")
	allNodes := append(nodesA, nodesB...)
	allPairs := append(pairsA, pairsB...)
	allPairs = append(allPairs, [2]domain.NodeID{"A1", "B1"})
	g := newTestGraph(allNodes, allPairs)

	c := &LouvainClusterer{Seed: 42}
	r1, _ := c.Cluster(context.Background(), g)
	r2, _ := c.Cluster(context.Background(), g)

	// 各クラスタのノード集合が一致するか比較
	toSets := func(clusters []domain.Cluster) [][]string {
		sets := make([][]string, 0, len(clusters))
		for _, cl := range clusters {
			ids := make([]string, len(cl.Nodes))
			for i, nid := range cl.Nodes {
				ids[i] = string(nid)
			}
			sort.Strings(ids)
			sets = append(sets, ids)
		}
		sort.Slice(sets, func(i, j int) bool { return sets[i][0] < sets[j][0] })
		return sets
	}

	s1, s2 := toSets(r1.Clusters), toSets(r2.Clusters)
	if len(s1) != len(s2) {
		t.Fatalf("cluster count differs: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if len(s1[i]) != len(s2[i]) {
			t.Errorf("cluster[%d] size differs", i)
			continue
		}
		for j := range s1[i] {
			if s1[i][j] != s2[i][j] {
				t.Errorf("cluster[%d][%d]: %q vs %q", i, j, s1[i][j], s2[i][j])
			}
		}
	}
}

func TestLouvainClusterer_AllNodesPreserved(t *testing.T) {
	nodesA, pairsA := buildK3("A")
	nodesB, pairsB := buildK3("B")
	allNodes := append(nodesA, nodesB...)
	allPairs := append(pairsA, pairsB...)
	g := newTestGraph(allNodes, allPairs)

	c := NewLouvainClusterer()
	result, err := c.Cluster(context.Background(), g)
	if err != nil {
		t.Fatalf("Cluster: %v", err)
	}

	// 全ノードがちょうど 1 つのクラスタに含まれることを確認
	seen := make(map[domain.NodeID]int)
	for _, cl := range result.Clusters {
		for _, nid := range cl.Nodes {
			seen[nid]++
		}
	}
	for _, n := range allNodes {
		if seen[n.ID] != 1 {
			t.Errorf("node %q appears %d times in clusters (want 1)", n.ID, seen[n.ID])
		}
	}
}

func TestLouvainClusterer_JSONRoundTrip(t *testing.T) {
	nodesA, pairsA := buildK3("A")
	g := newTestGraph(nodesA, pairsA)

	c := NewLouvainClusterer()
	result, err := c.Cluster(context.Background(), g)
	if err != nil {
		t.Fatalf("Cluster: %v", err)
	}

	b, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got domain.ClusterResult
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if len(got.Clusters) != len(result.Clusters) {
		t.Errorf("cluster count after round-trip: got %d, want %d", len(got.Clusters), len(result.Clusters))
	}
}
