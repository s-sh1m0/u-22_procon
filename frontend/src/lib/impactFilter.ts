import type { GraphResponse, GraphEdge } from '@/types/api'
import { isChanged } from '@/lib/nodeChange'

/**
 * 変更ノード（diff に関与）とその直接近傍 1-hop（呼び出し元・呼び出し先の両方向）
 * だけを残した部分グラフを返す。レビューでほぼ見ない無関係なノード / エッジを落とし、
 * 描画負荷と視覚ノイズを同時に下げるための純粋関数。
 *
 * クラスタ集約前の生グラフ（`data.graph`）に対して適用する想定。
 *
 * - エッジ: いずれかの端点が変更ノードなら残す
 * - interface メソッドノードは呼び出しの中継点として通り抜ける。変更ノードが呼ぶ interface
 *   メソッドは実装への implements エッジも、変更ノードを実装に持つ interface メソッドは
 *   呼び出し元からのエッジも残す（バックエンドの ImpactSubgraph と同じ範囲）。
 *   変更ノードが呼ぶ interface メソッドの別の呼び出し元のような兄弟は拾わない
 * - ノード: 変更ノード自身 + 残したエッジの端点
 * - 孤立した変更ノード（エッジを持たない）も残す
 * - clusters: 残ったノードで絞り込み、空になったクラスタは除去
 * - cycles / violations / pr はそのまま（端点が残らなければ描画側で無視される）
 *
 * 入力は破壊しない（新しい GraphResponse を返す）。
 */
export function filterToImpact(data: GraphResponse): GraphResponse {
  const changed = new Set<string>()
  const ifaces = new Set<string>()
  for (const n of data.graph.nodes) {
    if (isChanged(n)) changed.add(n.id)
    if (n.kind === 'interface_method') ifaces.add(n.id)
  }

  const calledIfaces = new Set<string>() // 変更ノードが呼ぶ interface メソッド
  const implByIfaces = new Set<string>() // 変更ノードを実装に持つ interface メソッド
  for (const e of data.graph.edges) {
    if (changed.has(e.from) && ifaces.has(e.to)) calledIfaces.add(e.to)
    if (changed.has(e.to) && ifaces.has(e.from)) implByIfaces.add(e.from)
  }

  const keepNodes = new Set<string>(changed)
  const keepEdges: GraphEdge[] = []
  for (const e of data.graph.edges) {
    const isImpl = e.kind === 'implements'
    if (
      changed.has(e.from) ||
      changed.has(e.to) ||
      (isImpl && calledIfaces.has(e.from)) ||
      (!isImpl && implByIfaces.has(e.to))
    ) {
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
