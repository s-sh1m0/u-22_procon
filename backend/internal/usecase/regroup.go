package usecase

import (
	"path"
	"sort"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// applyClusterMode は ClusterMode に応じて ClusterResult のクラスタ分割を上書きして返す。
// result が nil の場合はそのまま返す。
//   - Louvain: 保存済みクラスタ分割を保ち、ラベルだけ labelClusters で付け直す
//     （Cluster() は Label 空のクラスタを返す契約）。
//   - Package/File: result.Graph のノードを各キーで groupBy しラベルも付与する。
func applyClusterMode(result *domain.ClusterResult, mode domain.ClusterMode) *domain.ClusterResult {
	if result == nil {
		return result
	}
	switch mode {
	case domain.ClusterModeLouvain:
		return &domain.ClusterResult{
			Clusters:        labelClusters(result.Clusters, result.Graph),
			Graph:           result.Graph,
			Cycles:          result.Cycles,
			LayerViolations: result.LayerViolations,
		}
	case domain.ClusterModePackage, domain.ClusterModeFile:
		return &domain.ClusterResult{
			Clusters:        regroupBy(result.Graph, mode),
			Graph:           result.Graph,
			Cycles:          result.Cycles,
			LayerViolations: result.LayerViolations,
		}
	default:
		return result
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
