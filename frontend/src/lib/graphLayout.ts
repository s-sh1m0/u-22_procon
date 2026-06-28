import type { CSSProperties } from 'react'
import { MarkerType, type Edge, type EdgeMarker } from '@xyflow/react'
import ELK, { type ElkNode } from 'elkjs/lib/elk.bundled.js'
import type { GraphResponse, Cluster, GraphEdge, DiffStatus } from '@/types/api'
import type {
  AnyFlowNode,
  LayerKind,
  FunctionNodeData,
  ClusterGroupData,
  SuperClusterNodeData,
} from '@/types/graph'
import { getClusterColor } from './clusterColors'
import { inferLayer } from './layerInference'

export const NODE_W = 200
export const NODE_H = 60

const SUPER_W = 240
const SUPER_H = 90

// クラスタコンテナの内側余白。top はラベル表示分を広めに取る（ClusterGroup の
// ヘッダがこの領域に乗る）。子ノードはこの余白の内側に ELK が配置する。
const CLUSTER_PAD_TOP = 36
const CLUSTER_PAD = 16

// ELK レイヤードレイアウトのオプション。呼ぶ側→呼ばれる側を上→下に並べる。
// hierarchyHandling=INCLUDE_CHILDREN でクラスタ枠を跨ぐエッジも階層化対象にする。
const ROOT_LAYOUT_OPTIONS: Record<string, string> = {
  'elk.algorithm': 'layered',
  'elk.direction': 'DOWN',
  'elk.hierarchyHandling': 'INCLUDE_CHILDREN',
  'elk.layered.spacing.nodeNodeBetweenLayers': '64',
  'elk.spacing.nodeNode': '40',
  'elk.spacing.componentComponent': '64',
  'elk.separateConnectedComponents': 'true',
  // クラスタ内のノード順をなるべく元の並びに寄せ、再描画時の安定性を上げる。
  'elk.layered.considerModelOrder.strategy': 'NODES_AND_EDGES',
}

const CLUSTER_LAYOUT_OPTIONS: Record<string, string> = {
  'elk.padding': `[top=${CLUSTER_PAD_TOP},left=${CLUSTER_PAD},bottom=${CLUSTER_PAD},right=${CLUSTER_PAD}]`,
}

const elk = new ELK()

type InputNode = {
  id: string
  name: string
  package: string
  file: string
  line: number
  changed: boolean
  diffStatus: DiffStatus
}

export type LayoutResult = { nodes: AnyFlowNode[]; edges: Edge[] }

/** "layer:clusterId" 形式のキーを返す */
export function makeClusterKey(pkg: string, clusterId: number): string {
  return `${inferLayer(pkg)}:${clusterId}`
}

// 集約エッジの DiffStatus 優先順: added > removed > existing
function mergeEdgeStatus(a: DiffStatus, b: DiffStatus): DiffStatus {
  if (a === 'added' || b === 'added') return 'added'
  if (a === 'removed' || b === 'removed') return 'removed'
  return 'existing'
}

// 新規循環参照を構成するエッジの強調色（red-500）。diff の色より優先する。
const CYCLE_EDGE_COLOR = '#ef4444'

function edgeColorByStatus(status: DiffStatus): string {
  switch (status) {
    case 'added':
      return '#10b981'
    case 'removed':
      return '#a8a29e'
    case 'existing':
    default:
      return '#d6d3d1'
  }
}

function edgeStyleByStatus(status: DiffStatus): CSSProperties {
  const stroke = edgeColorByStatus(status)
  switch (status) {
    case 'added':
      return { stroke, strokeWidth: 2.5 }
    case 'removed':
      return { stroke, strokeWidth: 1.5, strokeDasharray: '6 4', opacity: 0.6 }
    case 'existing':
    default:
      return { stroke, strokeWidth: 1.5 }
  }
}

// 呼び出し方向（caller → callee）を示す矢印マーカー。色は stroke と揃える。
function edgeMarkerByStatus(status: DiffStatus): EdgeMarker {
  return { type: MarkerType.ArrowClosed, color: edgeColorByStatus(status), width: 16, height: 16 }
}

// 新規循環参照を構成するエッジのスタイル / マーカー（赤・太線）。
function cycleEdgeStyle(): CSSProperties {
  return { stroke: CYCLE_EDGE_COLOR, strokeWidth: 2.5 }
}
function cycleEdgeMarker(): EdgeMarker {
  return { type: MarkerType.ArrowClosed, color: CYCLE_EDGE_COLOR, width: 16, height: 16 }
}

// 展開クラスタ（コンテナ + 子関数ノード）/ 折りたたみクラスタ（スーパーノード）を
// 組み立てるための中間表現。ELK レイアウト後に位置を埋めて AnyFlowNode 化する。
type PendingCluster = {
  kind: 'cluster'
  data: ClusterGroupData
  children: { id: string; data: FunctionNodeData }[]
}
type PendingSuper = { kind: 'super'; data: SuperClusterNodeData }
type Pending = PendingCluster | PendingSuper

/**
 * グラフ全体を ELK の layered アルゴリズムで階層配置する。ノード/エッジの集約・
 * diff ステータス・循環強調のロジックはレイアウト非依存で、位置のみ ELK に委ねる。
 * クラスタ枠を compound ノードとして扱い、枠を跨ぐ呼び出し連鎖も上→下で読める。
 */
export async function layoutGraph(
  data: GraphResponse,
  expandedClusters: Set<string>,
): Promise<LayoutResult> {
  const inputNodes: InputNode[] = data.graph.nodes.map((n) => ({
    id: n.id,
    name: n.name,
    package: n.package,
    file: n.file,
    line: n.line,
    changed: n.changed,
    diffStatus: n.diff_status,
  }))
  const inputEdges: GraphEdge[] = data.graph.edges
  const clusters: Cluster[] = data.clusters

  // 新規循環参照（is_new）に属するノード集合と、同一循環内のノード対集合を作る。
  const newCycleSets = data.cycles.filter((c) => c.is_new).map((c) => new Set(c.nodes))
  const cycleNodeIds = new Set<string>()
  for (const s of newCycleSets) for (const id of s) cycleNodeIds.add(id)
  const isCycleEdge = (from: string, to: string): boolean =>
    newCycleSets.some((s) => s.has(from) && s.has(to))

  const nodeToCluster = new Map<string, number>()
  for (const c of clusters) {
    for (const nid of c.nodes) nodeToCluster.set(nid, c.id)
  }

  const nodeMap = new Map<string, InputNode>()
  for (const n of inputNodes) nodeMap.set(n.id, n)

  // (layer, clusterId) でノードをグルーピング
  const groups = new Map<string, InputNode[]>()
  for (const n of inputNodes) {
    const layer = inferLayer(n.package)
    const cid = nodeToCluster.get(n.id) ?? 0
    const key = `${layer}:${cid}`
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(n)
  }

  // エッジを集約。クラスタ折りたたみ時は端点をスーパーノードに付け替える。
  // DiffStatus は added > removed > existing で昇格。
  const edgeStatusMap = new Map<string, DiffStatus>()
  const edgeKeyToFromTo = new Map<string, { source: string; target: string }>()
  const edgeCycleMap = new Map<string, boolean>()

  for (const e of inputEdges) {
    const fromNode = nodeMap.get(e.from)
    const toNode = nodeMap.get(e.to)
    if (!fromNode || !toNode) continue

    const fromCid = nodeToCluster.get(e.from) ?? 0
    const toCid = nodeToCluster.get(e.to) ?? 0
    const fromKey = makeClusterKey(fromNode.package, fromCid)
    const toKey = makeClusterKey(toNode.package, toCid)

    const srcId = expandedClusters.has(fromKey) ? e.from : `super:${fromKey}`
    const dstId = expandedClusters.has(toKey) ? e.to : `super:${toKey}`

    if (srcId === dstId) continue

    const edgeKey = `${srcId}→${dstId}`
    const prev = edgeStatusMap.get(edgeKey)
    edgeStatusMap.set(edgeKey, prev ? mergeEdgeStatus(prev, e.status) : e.status)
    if (!edgeKeyToFromTo.has(edgeKey)) {
      edgeKeyToFromTo.set(edgeKey, { source: srcId, target: dstId })
    }
    if (isCycleEdge(e.from, e.to)) edgeCycleMap.set(edgeKey, true)
  }

  const flowEdges: Edge[] = []
  for (const [edgeKey, status] of edgeStatusMap.entries()) {
    const ft = edgeKeyToFromTo.get(edgeKey)!
    const inCycle = edgeCycleMap.get(edgeKey) ?? false
    flowEdges.push({
      id: `e:${edgeKey}`,
      source: ft.source,
      target: ft.target,
      style: inCycle ? cycleEdgeStyle() : edgeStyleByStatus(status),
      markerEnd: inCycle ? cycleEdgeMarker() : edgeMarkerByStatus(status),
      data: { diffStatus: status, inCycle },
    })
  }

  // ELK 入力（compound ノード）と、レイアウト後に位置を流し込むための登録簿を作る。
  const elkChildren: ElkNode[] = []
  const pending = new Map<string, Pending>()

  for (const [key, nodes] of groups.entries()) {
    const colonIdx = key.indexOf(':')
    const layer = key.slice(0, colonIdx) as LayerKind
    const cid = Number(key.slice(colonIdx + 1))
    const color = getClusterColor(cid)
    const clusterLabel = clusters.find((c) => c.id === cid)?.label ?? `Cluster ${cid}`

    if (!expandedClusters.has(key)) {
      // 折りたたみ: スーパーノード1個に集約
      const changedCount = nodes.reduce((acc, n) => acc + (n.changed ? 1 : 0), 0)
      const addedCount = nodes.reduce((acc, n) => acc + (n.diffStatus === 'added' ? 1 : 0), 0)
      const removedCount = nodes.reduce((acc, n) => acc + (n.diffStatus === 'removed' ? 1 : 0), 0)
      const superId = `super:${key}`

      elkChildren.push({ id: superId, width: SUPER_W, height: SUPER_H })
      pending.set(superId, {
        kind: 'super',
        data: {
          label: clusterLabel,
          clusterId: cid,
          clusterKey: key,
          clusterColorHex: color.hex,
          clusterColorSoft: color.soft,
          functionCount: nodes.length,
          changedCount,
          addedCount,
          removedCount,
          hasChanged: changedCount > 0 || addedCount > 0 || removedCount > 0,
        },
      })
    } else {
      // 展開: クラスタコンテナ + 子関数ノード
      const clusterId = `cluster:${key}`
      elkChildren.push({
        id: clusterId,
        layoutOptions: CLUSTER_LAYOUT_OPTIONS,
        children: nodes.map((n) => ({ id: n.id, width: NODE_W, height: NODE_H })),
      })
      pending.set(clusterId, {
        kind: 'cluster',
        data: {
          label: clusterLabel,
          clusterId: cid,
          clusterKey: key,
          clusterColorHex: color.hex,
          clusterColorSoft: color.soft,
        },
        children: nodes.map((n) => ({
          id: n.id,
          data: {
            label: n.name,
            packagePath: n.package,
            file: n.file,
            line: n.line,
            changed: n.changed,
            diffStatus: n.diffStatus,
            inCycle: cycleNodeIds.has(n.id),
            clusterId: cid,
            clusterColorHex: color.hex,
            layer,
          },
        })),
      })
    }
  }

  const elkGraph: ElkNode = {
    id: 'root',
    layoutOptions: ROOT_LAYOUT_OPTIONS,
    children: elkChildren,
    edges: flowEdges.map((e) => ({ id: e.id, sources: [e.source], targets: [e.target] })),
  }

  const laidOut = await elk.layout(elkGraph)

  // ELK の結果から React Flow ノードを組み立てる。親（コンテナ）は子より先に push する。
  const flowNodes: AnyFlowNode[] = []
  for (const child of laidOut.children ?? []) {
    const p = pending.get(child.id)
    if (!p) continue
    const x = child.x ?? 0
    const y = child.y ?? 0

    if (p.kind === 'super') {
      flowNodes.push({
        id: child.id,
        type: 'supercluster',
        position: { x, y },
        width: SUPER_W,
        height: SUPER_H,
        style: { width: SUPER_W, height: SUPER_H },
        data: p.data,
      })
    } else {
      const w = child.width ?? NODE_W
      const h = child.height ?? NODE_H
      flowNodes.push({
        id: child.id,
        type: 'cluster',
        position: { x, y },
        width: w,
        height: h,
        // onlyRenderVisibleElements の交差判定はトップレベルの width/height を読む。
        style: { width: w, height: h },
        draggable: false,
        selectable: true,
        zIndex: 0,
        data: p.data,
      })

      const childPos = new Map((child.children ?? []).map((c) => [c.id, c]))
      for (const cn of p.children) {
        const ec = childPos.get(cn.id)
        flowNodes.push({
          id: cn.id,
          type: 'function',
          parentId: child.id,
          extent: 'parent',
          // ELK の子座標は親コンテナの原点（左上）基準。React Flow の親子相対座標に一致。
          position: { x: ec?.x ?? 0, y: ec?.y ?? 0 },
          width: NODE_W,
          height: NODE_H,
          zIndex: 1,
          data: cn.data,
        })
      }
    }
  }

  return { nodes: flowNodes, edges: flowEdges }
}
