package analyzer

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// RefineChangedByContent は head グラフで Changed=true のノードのうち、base グラフに同じ ID があり
// 両側の関数ソース（Line〜EndLine 行）が一致するものを Changed=false にする。
//
// hunk の行範囲だけで判定すると、diff が共通の末尾行（`return nil` / `}` 等）を揃えた結果
// 中身の変わっていない隣の関数に +行がかかったり、patch が空でファイル単位にフォールバックしたりして、
// 無関係な関数まで changed になる。中身を比べて、実際に変わった関数だけを残す。
// Changed=false を true にすることはない（base ブランチ側だけの変更を拾わないため、hunk 判定との AND にする）。
//
// head を直接書き換える。headRoot / baseRoot は各グラフのノードの File が相対パスとする基点。
// 比較できないノード（base に同じ ID が無い・行範囲が不明・ファイルを読めない）は hunk の判定を維持する。
func RefineChangedByContent(head, base *domain.Graph, headRoot, baseRoot string) {
	if head == nil || base == nil {
		return
	}

	baseNodes := make(map[domain.NodeID]domain.Node, len(base.Nodes))
	for _, n := range base.Nodes {
		baseNodes[n.ID] = n
	}
	headSrc := newSourceCache(headRoot)
	baseSrc := newSourceCache(baseRoot)

	for i := range head.Nodes {
		hn := &head.Nodes[i]
		if !hn.Changed {
			continue
		}
		bn, ok := baseNodes[hn.ID]
		if !ok {
			continue
		}
		h, ok := headSrc.funcSource(*hn)
		if !ok {
			continue
		}
		b, ok := baseSrc.funcSource(bn)
		if !ok {
			continue
		}
		if h == b {
			hn.Changed = false
		}
	}
}

// sourceCache は 1 つの作業ツリーのファイルを行単位で読み、パスごとにキャッシュする。
type sourceCache struct {
	root  string
	lines map[string][]string // File → 行。読めなかったファイルは nil
}

func newSourceCache(root string) *sourceCache {
	return &sourceCache{root: root, lines: make(map[string][]string)}
}

// funcSource はノードの Line〜EndLine 行のソースを返す。
// 行範囲が不明・ファイルを読めない・行範囲がファイル外のときは false を返す。
func (c *sourceCache) funcSource(n domain.Node) (string, bool) {
	if n.File == "" || n.Line <= 0 || n.EndLine < n.Line {
		return "", false
	}
	lines, ok := c.lines[n.File]
	if !ok {
		path := n.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(c.root, path)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			// 比較できないだけなので hunk 判定のまま続行する（解析自体は失敗させない）。
			log.Printf("analyzer: refine changed: read %s: %v", path, err)
		} else {
			lines = strings.Split(string(content), "\n")
		}
		c.lines[n.File] = lines
	}
	if n.EndLine > len(lines) {
		return "", false
	}
	return strings.Join(lines[n.Line-1:n.EndLine], "\n"), true
}
