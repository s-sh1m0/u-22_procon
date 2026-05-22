import type { Edge } from '@xyflow/react'
import type { CSSProperties } from 'react'
import type { AnyFlowNode } from '@/types/graph'
import { type FocusDir, focusDirColor } from './graphFocus'

/**
 * 影響範囲（blast radius）モードの計算結果。
 * 変更ノード（origin）を起点に、呼び出しグラフを上流（caller）/ 下流（callee）へ
 * 推移的に辿った到達集合と、起点からのホップ距離を持つ。
 */
export type BlastResult = {
  /** origin + 上流 + 下流 のノード id 集合 */
  nodeIds: Set<string>
  /** 変更ノード（起点）の id 集合 */
  originIds: Set<string>
  /** 上流（呼び出し元）の id → 起点からのホップ距離(>=1) */
  upstream: Map<string, number>
  /** 下流（呼び出し先）の id → 起点からのホップ距離(>=1) */
  downstream: Map<string, number>
  /** エッジ id → 起点から見た向き（in = 上流側 / out = 下流側） */
  edgeDir: Map<string, FocusDir>
  /** エッジ id → 起点に近い側のホップ距離 */
  edgeHop: Map<string, number>
  originCount: number
  upstreamCount: number
  downstreamCount: number
}

function pushEdge(map: Map<string, Edge[]>, key: string, e: Edge): void {
  const arr = map.get(key)
  if (arr) arr.push(e)
  else map.set(key, [e])
}

/** flow ノードが変更（origin）かどうか。折りたたみ中の supercluster も対象にする。 */
function isOriginNode(n: AnyFlowNode): boolean {
  if (n.type === 'function') return n.data.changed || n.data.diffStatus !== 'existing'
  if (n.type === 'supercluster') return n.data.hasChanged
  return false
}

/**
 * 全変更ノードを起点に、上流（caller）/ 下流（callee）双方へ多ホップで波及範囲を算出する。
 * エッジは source = caller, target = callee（呼び出し方向）。
 * - 上流 = 変更ノードを呼ぶ側（変更の影響を受けるため要確認）
 * - 下流 = 変更ノードが呼ぶ側（変更が依存している先）
 *
 * 現在のレイアウト（クラスタ折りたたみ反映済み）の nodes / edges をそのまま受け取るため、
 * 折りたたみ状態に整合した波及範囲になる。
 */
export function computeBlastRadius(nodes: AnyFlowNode[], edges: Edge[]): BlastResult {
  const originIds = new Set<string>()
  for (const n of nodes) {
    if (isOriginNode(n)) originIds.add(n.id)
  }

  // 隣接リスト: incoming[target] で caller を、outgoing[source] で callee を引く
  const incoming = new Map<string, Edge[]>()
  const outgoing = new Map<string, Edge[]>()
  for (const e of edges) {
    pushEdge(incoming, e.target, e)
    pushEdge(outgoing, e.source, e)
  }

  const edgeDir = new Map<string, FocusDir>()
  const edgeHop = new Map<string, number>()

  const upstream = traverse(originIds, incoming, (e) => e.source, 'in', edgeDir, edgeHop)
  const downstream = traverse(originIds, outgoing, (e) => e.target, 'out', edgeDir, edgeHop)

  const nodeIds = new Set<string>(originIds)
  for (const id of upstream.keys()) nodeIds.add(id)
  for (const id of downstream.keys()) nodeIds.add(id)

  return {
    nodeIds,
    originIds,
    upstream,
    downstream,
    edgeDir,
    edgeHop,
    originCount: originIds.size,
    upstreamCount: upstream.size,
    downstreamCount: downstream.size,
  }
}

/**
 * origin から一方向に BFS する共通処理。
 * adjacency は探索方向の隣接（上流なら incoming、下流なら outgoing）。
 * nextOf はエッジから次に進むノード id を返す。
 * 探索したエッジには方向 dir と起点側ホップを記録する（未設定のもののみ）。
 */
function traverse(
  originIds: Set<string>,
  adjacency: Map<string, Edge[]>,
  nextOf: (e: Edge) => string,
  dir: FocusDir,
  edgeDir: Map<string, FocusDir>,
  edgeHop: Map<string, number>,
): Map<string, number> {
  const dist = new Map<string, number>()
  for (const id of originIds) dist.set(id, 0)
  const reached = new Map<string, number>()
  const queue = [...originIds]
  while (queue.length) {
    const cur = queue.shift()!
    const curDist = dist.get(cur)!
    for (const e of adjacency.get(cur) ?? []) {
      if (!edgeDir.has(e.id)) {
        edgeDir.set(e.id, dir)
        edgeHop.set(e.id, curDist)
      }
      const next = nextOf(e)
      if (!dist.has(next)) {
        dist.set(next, curDist + 1)
        reached.set(next, curDist + 1)
        queue.push(next)
      }
    }
  }
  return reached
}

const MIN_BLAST_EDGE_OPACITY = 0.35
const MIN_BLAST_NODE_OPACITY = 0.5

/** 影響範囲エッジのスタイル。起点から遠いほど細く・薄くしてホップ距離を表現する。 */
export function blastEdgeStyle(dir: FocusDir, hop: number): CSSProperties {
  return {
    stroke: focusDirColor(dir),
    strokeWidth: Math.max(1.2, 2.5 - hop * 0.35),
    opacity: Math.max(MIN_BLAST_EDGE_OPACITY, 1 - hop * 0.18),
  }
}

/** 影響範囲ノードの opacity。起点から遠いほど薄くする。 */
export function blastNodeOpacity(hop: number): number {
  return Math.max(MIN_BLAST_NODE_OPACITY, 1 - hop * 0.14)
}
