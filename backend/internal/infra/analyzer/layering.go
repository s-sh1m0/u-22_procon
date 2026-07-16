package analyzer

import (
	"sort"
	"strings"

	"github.com/s-sh1m0/u-22_procon/backend/internal/domain"
)

// layerRank はクリーンアーキテクチャ上のレイヤー深さを返す（小さいほど内側）。
// rank が小さいレイヤーから大きいレイヤーへの呼び出しは「依存方向の逆転」。
// 判定対象外（pkg / 外部依存など）は rank=-1 を返す。
//
// レイヤー判定はフロントの inferLayer（frontend/src/lib/layerInference.ts）と
// 整合させる。usecase と domain はどちらも「内側（domain）」として 1 つに束ねる
// （usecase → infra も domain → infra と同じく禁止依存のため）。
func layerRank(pkg string) (name string, rank int) {
	switch {
	case strings.Contains(pkg, "internal/domain"), strings.Contains(pkg, "internal/usecase"):
		return "domain", 0
	case strings.Contains(pkg, "internal/infra"):
		return "infra", 1
	case strings.Contains(pkg, "internal/api"), strings.Contains(pkg, "cmd/"):
		return "ui", 2
	default:
		return "", -1
	}
}

// DetectLayeringViolations は差分グラフ上で依存方向の逆転を検出する。
// 内側レイヤー（rank 小）から外側レイヤー（rank 大）への呼び出しエッジを違反とみなす。
// エッジの Status が added のものを IsNew=true とする（PR で新規に持ち込まれた依存）。
//
// 戻り値は From, To の昇順でソート済み（テストの再現性のため）。
func DetectLayeringViolations(g *domain.Graph) []domain.LayerViolation {
	if g == nil || len(g.Edges) == 0 {
		return nil
	}

	pkgOf := make(map[domain.NodeID]string, len(g.Nodes))
	for _, n := range g.Nodes {
		pkgOf[n.ID] = n.Package
	}

	out := make([]domain.LayerViolation, 0)
	for _, e := range g.Edges {
		fromLayer, fromRank := layerRank(pkgOf[e.From])
		toLayer, toRank := layerRank(pkgOf[e.To])
		if fromRank < 0 || toRank < 0 {
			continue
		}
		// 内側 → 外側 のみ違反。同一レイヤー・外側 → 内側（infra → domain 等）は正常。
		if fromRank >= toRank {
			continue
		}
		out = append(out, domain.LayerViolation{
			From:      e.From,
			To:        e.To,
			FromLayer: fromLayer,
			ToLayer:   toLayer,
			IsNew:     e.Status == domain.DiffStatusAdded,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}
