import { MarkerType, type Edge, type EdgeMarker } from '@xyflow/react'
import type { CSSProperties } from 'react'

/** 選択ノードから見たエッジの向き。in = 呼び出し元→選択, out = 選択→呼び出し先 */
export type FocusDir = 'in' | 'out'

// 呼び出し方向を示す色。呼び出し元(upstream caller) と呼び出し先(downstream callee) を区別する。
export const FOCUS_CALLER_COLOR = '#3b82f6' // blue: 呼び出し元（選択ノードを呼ぶ側）
export const FOCUS_CALLEE_COLOR = '#f97316' // orange: 呼び出し先（選択ノードが呼ぶ側）

export function focusDirColor(dir: FocusDir): string {
  return dir === 'in' ? FOCUS_CALLER_COLOR : FOCUS_CALLEE_COLOR
}

export type FocusResult = {
  /** 選択ノード + 直接 caller/callee の id 集合 */
  nodeIds: Set<string>
  /** エッジ id → 選択ノードから見た向き */
  edgeDir: Map<string, FocusDir>
  callerCount: number
  calleeCount: number
}

/**
 * 選択ノードの 1-hop 近傍（直接の caller / callee）を flow エッジから算出する。
 * source(下) → target(上) が caller → callee。つまり target === selected なら相手は caller。
 */
export function computeFocus(edges: Edge[], selectedNodeId: string): FocusResult {
  const nodeIds = new Set<string>([selectedNodeId])
  const edgeDir = new Map<string, FocusDir>()
  let callerCount = 0
  let calleeCount = 0
  for (const e of edges) {
    if (e.source === selectedNodeId) {
      nodeIds.add(e.target)
      edgeDir.set(e.id, 'out')
      calleeCount++
    } else if (e.target === selectedNodeId) {
      nodeIds.add(e.source)
      edgeDir.set(e.id, 'in')
      callerCount++
    }
  }
  return { nodeIds, edgeDir, callerCount, calleeCount }
}

export function focusEdgeStyle(dir: FocusDir): CSSProperties {
  return { stroke: focusDirColor(dir), strokeWidth: 2.5, opacity: 1 }
}

export function focusEdgeMarker(dir: FocusDir): EdgeMarker {
  return { type: MarkerType.ArrowClosed, color: focusDirColor(dir), width: 18, height: 18 }
}

// フォーカス外のエッジ。方向色は消し、強く減光して背景化する。
export const DIMMED_EDGE_STYLE: CSSProperties = { stroke: '#e7e5e4', strokeWidth: 1, opacity: 0.12 }

export const FOCUSED_OPACITY = 1
export const DIMMED_OPACITY = 0.18
