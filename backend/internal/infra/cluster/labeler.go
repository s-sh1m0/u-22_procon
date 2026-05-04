package cluster

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// labelCluster はクラスタを構成するノード群から人間に読めるラベルを生成する。
//
// 決定順:
//  1. ノードの File パスの最長共通プレフィックス（/ 境界に丸める）
//  2. 最頻出 Package（同数ならアルファベット順）
//  3. 最頻出 filepath.Base(File)
//  4. fallback: "cluster N"
//
// Changed=true ノードのうちクラスタ内エッジ次数最大の関数名を " (#Name)" で付加する。
func labelCluster(idx int, nodes []domain.Node, edges []domain.Edge) string {
	if len(nodes) == 0 {
		return fmt.Sprintf("cluster %d", idx)
	}

	base := baseLabel(idx, nodes)

	rep := pickRepresentativeChangedNode(nodes, edges)
	if rep != "" {
		return base + " (#" + rep + ")"
	}
	return base
}

func baseLabel(idx int, nodes []domain.Node) string {
	// 1. 共通ファイルパスプレフィックス
	files := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.File != "" {
			files = append(files, n.File)
		}
	}
	if len(files) > 0 {
		prefix := commonPathPrefix(files)
		if prefix != "" {
			return prefix
		}
	}

	// 2/3. 最頻出パッケージ vs 最頻出 basename を比較し、多い方を採用。
	// 同頻度ならパッケージを優先する。
	pkgSlice := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.Package != "" {
			pkgSlice = append(pkgSlice, n.Package)
		}
	}
	baseSlice := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.File != "" {
			baseSlice = append(baseSlice, filepath.Base(n.File))
		}
	}
	pkg, pkgFreq := mostCommonWithFreq(pkgSlice)
	base, baseFreq := mostCommonWithFreq(baseSlice)
	if baseFreq > pkgFreq && base != "" {
		return base
	}
	if pkg != "" {
		return pkg
	}
	if base != "" {
		return base
	}

	return fmt.Sprintf("cluster %d", idx)
}

// commonPathPrefix はファイルパスのスラッシュ区切り境界に丸めた共通プレフィックスを返す。
func commonPathPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	prefix := paths[0]
	for _, p := range paths[1:] {
		prefix = commonString(prefix, p)
		if prefix == "" {
			return ""
		}
	}
	// / 境界に丸める
	if idx := strings.LastIndex(prefix, "/"); idx > 0 {
		prefix = prefix[:idx]
	} else {
		return ""
	}
	return prefix
}

func commonString(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:n]
}

// mostCommonWithFreq は最頻出文字列とその出現回数を返す。空スライスなら ("", 0)。
func mostCommonWithFreq(vals []string) (string, int) {
	if len(vals) == 0 {
		return "", 0
	}
	freq := make(map[string]int, len(vals))
	for _, v := range vals {
		freq[v]++
	}
	keys := make([]string, 0, len(freq))
	for k := range freq {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	best := keys[0]
	for _, k := range keys[1:] {
		if freq[k] > freq[best] {
			best = k
		}
	}
	return best, freq[best]
}

// pickRepresentativeChangedNode はクラスタ内の Changed ノードのうち、
// クラスタ内エッジ次数が最大のものの Name を返す。候補なしなら ""。
func pickRepresentativeChangedNode(nodes []domain.Node, edges []domain.Edge) string {
	nodeSet := make(map[domain.NodeID]struct{}, len(nodes))
	for _, n := range nodes {
		nodeSet[n.ID] = struct{}{}
	}

	degree := make(map[domain.NodeID]int)
	for _, e := range edges {
		_, fromIn := nodeSet[e.From]
		_, toIn := nodeSet[e.To]
		if fromIn && toIn {
			degree[e.From]++
			degree[e.To]++
		}
	}

	var bestID domain.NodeID
	bestDeg := -1
	for _, n := range nodes {
		if !n.Changed {
			continue
		}
		d := degree[n.ID]
		if d > bestDeg || (d == bestDeg && (bestDeg == -1 || n.Name < string(bestID))) {
			bestDeg = d
			bestID = n.ID
		}
	}

	if bestDeg == -1 {
		return ""
	}
	for _, n := range nodes {
		if n.ID == bestID {
			return n.Name
		}
	}
	return ""
}
