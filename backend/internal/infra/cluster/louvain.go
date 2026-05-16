package cluster

import (
	"context"
	"math/rand/v2"
	"sort"

	"gonum.org/v1/gonum/graph/community"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// Clusterer は呼び出しグラフをクラスタ分割する抽象。
type Clusterer interface {
	Cluster(ctx context.Context, g *domain.Graph) (*domain.ClusterResult, error)
}

// LouvainClusterer は gonum の community.Modularize を用いた Louvain 法実装。
type LouvainClusterer struct {
	// Resolution は modularity の解像度パラメータ。
	// 0 以下の場合は resolveResolution() でノード数に応じた値を自動決定する。
	Resolution float64

	// Seed は乱数シード。0 の場合は 1 を使用（決定的出力のため）。
	Seed int64
}

// NewLouvainClusterer はデフォルト設定の LouvainClusterer を返す。
// Resolution=0（自動調整）、Seed=1（決定的）。
func NewLouvainClusterer() *LouvainClusterer {
	return &LouvainClusterer{}
}

// Cluster は domain.Graph を Louvain 法でクラスタリングし domain.ClusterResult を返す。
// g が nil またはノードを持たない場合は空の ClusterResult を返す。
//
// 返却される Cluster の Label は空のままにする。表示用ラベルは usecase 層が読み出し時に付与する
// （Package/File モードと対称にするため。`.claude/rules/backend-architecture.md` §4 参照）。
func (c *LouvainClusterer) Cluster(ctx context.Context, g *domain.Graph) (*domain.ClusterResult, error) {
	if g == nil || len(g.Nodes) == 0 {
		return &domain.ClusterResult{}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	adapter := newGonumAdapter(g)
	resolution := resolveResolution(len(g.Nodes), c.Resolution)

	seed := c.Seed
	if seed == 0 {
		seed = 1
	}
	src := rand.NewPCG(uint64(seed), 0)

	reduced := community.Modularize(adapter.g, resolution, src)
	communities := reduced.Communities()

	// サイズ降順でソートして安定した ID を付与する
	sort.Slice(communities, func(i, j int) bool {
		return len(communities[i]) > len(communities[j])
	})

	clusters := make([]domain.Cluster, 0, len(communities))
	for i, comm := range communities {
		if len(comm) == 0 {
			continue
		}

		nodeIDs := make([]domain.NodeID, 0, len(comm))
		for _, gNode := range comm {
			domainID, ok := adapter.idToNode[gNode.ID()]
			if !ok {
				continue
			}
			nodeIDs = append(nodeIDs, domainID)
		}

		clusters = append(clusters, domain.Cluster{
			ID:    i,
			Nodes: nodeIDs,
		})
	}

	return &domain.ClusterResult{
		Clusters: clusters,
		Graph:    *g,
	}, nil
}

// resolveResolution はノード数と override から Louvain の resolution を決定する。
// override > 0 の場合はそれをそのまま使用する。
// それ以外はノード数バケットで段階的に調整する（小さい PR は過分割を避け、大きい PR は細分化を促す）。
func resolveResolution(nodeCount int, override float64) float64 {
	if override > 0 {
		return override
	}
	switch {
	case nodeCount < 30:
		return 0.7
	case nodeCount < 100:
		return 1.0
	case nodeCount < 300:
		return 1.4
	default:
		return 1.8
	}
}
