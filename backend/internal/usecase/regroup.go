package usecase

import (
	"path"
	"sort"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// applyClusterMode は ClusterMode に応じて ClusterResult のクラスタ分割を上書きして返す。
// mode が ClusterModeLouvain または result が nil の場合はそのまま返す。
// Package/File モードは result.Graph のノードを各キーで groupBy する。
func applyClusterMode(result *domain.ClusterResult, mode domain.ClusterMode) *domain.ClusterResult {
	if result == nil || mode == domain.ClusterModeLouvain {
		return result
	}
	return &domain.ClusterResult{
		Clusters: regroupBy(result.Graph, mode),
		Graph:    result.Graph,
		Cycles:   result.Cycles,
	}
}

// regroupBy は mode に対応するノード属性でグラフのノードを groupBy し、
// サイズ降順・キー昇順で安定したクラスタ列を返す。
// 未対応の mode は空スライスを返す。
func regroupBy(g domain.Graph, mode domain.ClusterMode) []domain.Cluster {
	var keyOf func(n domain.Node) string
	var labelOf func(key string) string

	switch mode {
	case domain.ClusterModePackage:
		keyOf = func(n domain.Node) string { return n.Package }
		labelOf = func(key string) string {
			if key == "" {
				return "(no package)"
			}
			return path.Base(key)
		}
	case domain.ClusterModeFile:
		keyOf = func(n domain.Node) string { return n.File }
		labelOf = func(key string) string {
			if key == "" {
				return "(no file)"
			}
			return path.Base(key)
		}
	default:
		return []domain.Cluster{}
	}

	groups := make(map[string][]domain.NodeID)
	for _, n := range g.Nodes {
		k := keyOf(n)
		groups[k] = append(groups[k], n.ID)
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		li, lj := len(groups[keys[i]]), len(groups[keys[j]])
		if li != lj {
			return li > lj
		}
		return keys[i] < keys[j]
	})

	clusters := make([]domain.Cluster, 0, len(keys))
	for i, k := range keys {
		clusters = append(clusters, domain.Cluster{
			ID:    i,
			Label: labelOf(k),
			Nodes: groups[k],
		})
	}
	return clusters
}
