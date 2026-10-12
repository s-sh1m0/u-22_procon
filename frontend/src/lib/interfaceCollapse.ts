import type { GraphResponse } from '@/types/api'
import { isChanged } from '@/lib/nodeChange'

export type CollapsedGraph = {
  /** 折りたたんだ実装ノード・implements エッジを除いたグラフ */
  data: GraphResponse
  /** interface メソッドノード ID → 実装の数（折りたたむ前のグラフで数えたもの） */
  implCounts: Map<string, number>
}

/**
 * interface メソッドの実装を、その interface メソッドが展開されるまで隠した部分グラフを返す。
 * interface 経由の呼び出しは呼び出し元 → interface メソッド → 実装（implements エッジ）の形で
 * 届くので、実装を全部並べると 1 回の呼び出しが実装の数だけ広がって見える。既定では隠し、
 * interface メソッドノードのクリックで展開する（段階的に見せる）。
 *
 * - 隠すノード: 変更ノードではなく、implements エッジの先になっていて、その interface メソッドが
 *   どれも展開されておらず、呼び出しエッジで呼ばれていないもの（interface 経由でしか届かない実装）
 * - 隠すエッジ: 隠したノードにつながるエッジと、展開されていない interface メソッドからの
 *   implements エッジ。ただし実装が変更ノードなら、どの interface から届くかを示すため残す
 * - clusters: 残ったノードで絞り込み、空になったクラスタは除去する
 *
 * 入力は破壊しない（新しい GraphResponse を返す）。
 */
export function collapseImplementations(
  data: GraphResponse,
  expanded: Set<string>,
): CollapsedGraph {
  const implCounts = new Map<string, number>()
  const ifacesOf = new Map<string, string[]>() // 実装 → それを実装に持つ interface メソッド
  const called = new Set<string>() // 呼び出しエッジで呼ばれているノード
  for (const e of data.graph.edges) {
    if (e.kind === 'implements') {
      implCounts.set(e.from, (implCounts.get(e.from) ?? 0) + 1)
      const ifaces = ifacesOf.get(e.to)
      if (ifaces) ifaces.push(e.from)
      else ifacesOf.set(e.to, [e.from])
    } else {
      called.add(e.to)
    }
  }

  const changed = new Set<string>()
  for (const n of data.graph.nodes) {
    if (isChanged(n)) changed.add(n.id)
  }

  const hidden = new Set<string>()
  for (const [id, ifaces] of ifacesOf) {
    if (changed.has(id) || called.has(id)) continue
    if (ifaces.some((i) => expanded.has(i))) continue
    hidden.add(id)
  }

  const edges = data.graph.edges.filter((e) => {
    if (hidden.has(e.from) || hidden.has(e.to)) return false
    return e.kind !== 'implements' || expanded.has(e.from) || changed.has(e.to)
  })
  const nodes = data.graph.nodes.filter((n) => !hidden.has(n.id))
  const clusters = data.clusters
    .map((c) => ({ ...c, nodes: c.nodes.filter((id) => !hidden.has(id)) }))
    .filter((c) => c.nodes.length > 0)

  return { data: { ...data, clusters, graph: { nodes, edges } }, implCounts }
}

/** interface メソッド ifaceId の実装ノード ID（implements エッジの先）を返す。 */
export function implementationIds(data: GraphResponse, ifaceId: string): string[] {
  return data.graph.edges
    .filter((e) => e.kind === 'implements' && e.from === ifaceId)
    .map((e) => e.to)
}
