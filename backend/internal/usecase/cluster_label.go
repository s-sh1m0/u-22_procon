package usecase

import (
	"fmt"
	"path"
	"unicode"
	"unicode/utf8"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// labelClusters は clusters の各クラスタに対し graph の情報からラベルを計算し、
// Label を差し替えた新しいスライスを返す。入力の clusters は変更しない。
func labelClusters(clusters []domain.Cluster, g domain.Graph) []domain.Cluster {
	if len(clusters) == 0 {
		return clusters
	}
	nodeByID := make(map[domain.NodeID]domain.Node, len(g.Nodes))
	for _, n := range g.Nodes {
		nodeByID[n.ID] = n
	}
	// 入次数はグラフ全体で一度だけ計算し、全クラスタで使い回す（O(E)）。
	// クラスタ毎に再計算すると O(C×E) になるため。
	inDeg := computeInDegree(g.Edges)
	out := make([]domain.Cluster, len(clusters))
	for i, c := range clusters {
		clusterNodes := make([]domain.Node, 0, len(c.Nodes))
		for _, id := range c.Nodes {
			if n, ok := nodeByID[id]; ok {
				clusterNodes = append(clusterNodes, n)
			}
		}
		labeled := c
		labeled.Label = labelCluster(c.ID, clusterNodes, inDeg)
		out[i] = labeled
	}
	return out
}

// labelCluster はクラスタを構成するノード群から人間に読めるラベルを生成する。
//
// 形式: "<短縮パッケージ名>.<代表関数名>" （例: "usecase.AnalyzePR"）
//
// 代表関数は以下の優先度で絞り込み、各段階で空集合になったらフィルタを適用しない:
//  1. Changed=true のノードを優先する
//  2. 公開関数（先頭大文字）を優先する
//  3. グラフ全体での入次数が最大のノードを選ぶ
//  4. tie-break は名前の昇順
//
// パッケージ名・関数名のいずれも欠けている場合は "cluster N" にフォールバックする。
func labelCluster(idx int, nodes []domain.Node, inDeg map[domain.NodeID]int) string {
	if len(nodes) == 0 {
		return fmt.Sprintf("cluster %d", idx)
	}

	rep := pickRepresentative(nodes, inDeg)
	short := shortPackage(rep.Package)

	switch {
	case short != "" && rep.Name != "":
		return short + "." + rep.Name
	case short != "":
		return short
	case rep.Name != "":
		return rep.Name
	default:
		return fmt.Sprintf("cluster %d", idx)
	}
}

// pickRepresentative はクラスタから代表ノードを選ぶ。len(nodes) > 0 を前提とする。
// inDeg はグラフ全体の入次数マップ（computeInDegree の結果）を想定する。
func pickRepresentative(nodes []domain.Node, inDeg map[domain.NodeID]int) domain.Node {
	candidates := filterNodes(nodes, func(n domain.Node) bool { return n.Changed })
	if len(candidates) == 0 {
		candidates = nodes
	}

	if exported := filterNodes(candidates, func(n domain.Node) bool { return isExported(n.Name) }); len(exported) > 0 {
		candidates = exported
	}

	best := candidates[0]
	for _, n := range candidates[1:] {
		bd, nd := inDeg[best.ID], inDeg[n.ID]
		if nd > bd || (nd == bd && n.Name < best.Name) {
			best = n
		}
	}
	return best
}

// computeInDegree はグラフ全体の各ノードの入次数を返す。
// edges は cluster を跨ぐ呼び出しも含むグラフ全体のエッジを想定する。
func computeInDegree(edges []domain.Edge) map[domain.NodeID]int {
	deg := make(map[domain.NodeID]int, len(edges))
	for _, e := range edges {
		deg[e.To]++
	}
	return deg
}

func filterNodes(nodes []domain.Node, pred func(domain.Node) bool) []domain.Node {
	out := make([]domain.Node, 0, len(nodes))
	for _, n := range nodes {
		if pred(n) {
			out = append(out, n)
		}
	}
	return out
}

// shortPackage はパッケージパスの末尾セグメントを返す。
// 例: "github.com/foo/bar/usecase" -> "usecase"
func shortPackage(pkg string) string {
	if pkg == "" {
		return ""
	}
	return path.Base(pkg)
}

// isExported は Go のエクスポート規則（先頭大文字）に従って公開関数か判定する。
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
