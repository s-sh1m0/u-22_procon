import type { Cycle, DiffStatus, GraphEdge } from '@/types/api'

// レビュー優先度。レビュアーがどのノードを重点的に見るべきかの指標。
export type ReviewPriority = 'high' | 'medium' | 'low'

// 「高」と判定する被呼び出し数（in-degree）のしきい値。
// これ以上の呼び出し元を持つ変更ノードは影響範囲が広いとみなす。
export const HIGH_IN_DEGREE = 3

// computeInDegree は各ノードの被呼び出し数（呼び出し元の数）を集計する。
export function computeInDegree(edges: GraphEdge[]): Map<string, number> {
  const m = new Map<string, number>()
  for (const e of edges) m.set(e.to, (m.get(e.to) ?? 0) + 1)
  return m
}

// newCycleNodeIds は新規循環参照（is_new）に属する全ノード ID 集合を返す。
export function newCycleNodeIds(cycles: Cycle[]): Set<string> {
  const s = new Set<string>()
  for (const c of cycles) {
    if (!c.is_new) continue
    for (const n of c.nodes) s.add(n)
  }
  return s
}

// computeReviewPriority はノード単体の優先度を決める。
//
// 高: 新規循環に含まれる、または PR で変更され被呼び出し数が多い（影響範囲が広い）
// 中: PR で変更されたが上記に当たらない
// 低: PR で変更されていない（レビューの主対象ではない文脈ノード）
export function computeReviewPriority(
  node: { changed: boolean; diff_status: DiffStatus },
  inDegree: number,
  inNewCycle: boolean,
): ReviewPriority {
  if (inNewCycle) return 'high'
  const touched = node.changed || node.diff_status !== 'existing'
  if (!touched) return 'low'
  if (inDegree >= HIGH_IN_DEGREE) return 'high'
  return 'medium'
}
