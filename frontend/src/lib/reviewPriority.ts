import type { DiffStatus, GraphEdge } from '@/types/api'
import { isChanged } from '@/lib/nodeChange'

// レビュー優先度。レビュアーがどのノードを重点的に見るべきかの指標。
export type ReviewPriority = 'high' | 'medium' | 'low'

// 「高」と判定する被呼び出し数（in-degree）のしきい値。
// これ以上の呼び出し元を持つ変更ノードは影響範囲が広いとみなす。
export const HIGH_IN_DEGREE = 3

// computeInDegree は各ノードの被呼び出し数（呼び出し元の数）を集計する。
// interface 経由の呼び出しは呼び出し元 → interface メソッド → 実装（implements）の形なので、
// 実装には interface メソッドの被呼び出し数を足す（直接の呼び出しと同じく影響範囲として数える）。
export function computeInDegree(edges: GraphEdge[]): Map<string, number> {
  const m = new Map<string, number>()
  for (const e of edges) {
    if (e.kind !== 'implements') m.set(e.to, (m.get(e.to) ?? 0) + 1)
  }
  const viaIface = new Map<string, number>()
  for (const e of edges) {
    if (e.kind === 'implements')
      viaIface.set(e.to, (viaIface.get(e.to) ?? 0) + (m.get(e.from) ?? 0))
  }
  for (const [id, n] of viaIface) m.set(id, (m.get(id) ?? 0) + n)
  return m
}

// computeReviewPriority はノード単体の優先度を決める。
//
// 高: PR で変更され被呼び出し数が多い（影響範囲が広い）
// 中: PR で変更されたが上記に当たらない
// 低: PR で変更されていない（レビューの主対象ではない文脈ノード）
export function computeReviewPriority(
  node: { changed: boolean; diff_status: DiffStatus },
  inDegree: number,
): ReviewPriority {
  if (!isChanged(node)) return 'low'
  if (inDegree >= HIGH_IN_DEGREE) return 'high'
  return 'medium'
}
