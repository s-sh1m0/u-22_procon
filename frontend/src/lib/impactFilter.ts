import type { GraphResponse, GraphEdge, GraphNode } from '@/types/api'

/** ノードが diff に関与しているか（追加 / 削除 / 変更のいずれか）。 */
function isChanged(n: GraphNode): boolean {
  return n.changed || n.diff_status !== 'existing'
}

/**
 * 変更ノード（diff に関与）とその直接近傍 1-hop（呼び出し元・呼び出し先の両方向）
 * だけを残した部分グラフを返す。レビューでほぼ見ない無関係なノード / エッジを落とし、
 * 描画負荷と視覚ノイズを同時に下げるための純粋関数。
 *
 * クラスタ集約前の生グラフ（`data.graph`）に対して適用する想定。
 *
 * - エッジ: いずれかの端点が変更ノードなら残す
 * - ノード: 変更ノード自身 + 残したエッジの端点
 * - 孤立した変更ノード（エッジを持たない）も残す
 * - clusters: 残ったノードで絞り込み、空になったクラスタは除去
 * - cycles / violations / pr はそのまま（端点が残らなければ描画側で無視される）
 *
 * 入力は破壊しない（新しい GraphResponse を返す）。
 */
export function filterToImpact(data: GraphResponse): GraphResponse {
  const changed = new Set<string>()
  for (const n of data.graph.nodes) {
    if (isChanged(n)) changed.add(n.id)
  }

  const keepNodes = new Set<string>(changed)
  const keepEdges: GraphEdge[] = []
  for (const e of data.graph.edges) {
    if (changed.has(e.from) || changed.has(e.to)) {
      keepEdges.push(e)
      keepNodes.add(e.from)
      keepNodes.add(e.to)
    }
  }

  const nodes = data.graph.nodes.filter((n) => keepNodes.has(n.id))
  const clusters = data.clusters
    .map((c) => ({ ...c, nodes: c.nodes.filter((id) => keepNodes.has(id)) }))
    .filter((c) => c.nodes.length > 0)

  return {
    ...data,
    clusters,
    graph: { nodes, edges: keepEdges },
  }
}
