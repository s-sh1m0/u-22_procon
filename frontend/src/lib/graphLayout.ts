import type { CSSProperties } from 'react'
import { MarkerType, type Edge, type EdgeMarker } from '@xyflow/react'
import ELKConstructor, { type ELK, type ElkNode } from 'elkjs/lib/elk-api.js'
import elkWorkerUrl from 'elkjs/lib/elk-worker.min.js?url'
import type { GraphResponse, Cluster, GraphEdge, DiffStatus, EdgeKind, NodeKind } from '@/types/api'
import type {
  AnyFlowNode,
  LayerKind,
  FunctionNodeData,
  ClusterGroupData,
  SuperClusterNodeData,
} from '@/types/graph'
import { getClusterColor } from './clusterColors'
import { inferLayer } from './layerInference'
import { nodeDisplayName } from './nodeLabel'
import { IMPLEMENTS_EDGE_DASH } from './graphFocus'

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
// considerModelOrder / separateConnectedComponents は付けない（計算コストがほぼ倍増し、
// 「すべて展開」時に体感できるレベルで重くなるため。実測で確認済み）。
const ROOT_LAYOUT_OPTIONS: Record<string, string> = {
  'elk.algorithm': 'layered',
  'elk.direction': 'DOWN',
  'elk.hierarchyHandling': 'INCLUDE_CHILDREN',
  'elk.layered.spacing.nodeNodeBetweenLayers': '64',
  'elk.spacing.nodeNode': '40',
  'elk.spacing.componentComponent': '64',
}

const CLUSTER_LAYOUT_OPTIONS: Record<string, string> = {
  'elk.padding': `[top=${CLUSTER_PAD_TOP},left=${CLUSTER_PAD},bottom=${CLUSTER_PAD},right=${CLUSTER_PAD}]`,
}

// ELK は Web Worker で実行する（メインスレッドを塞がないため）。layered レイアウトは
// ノード数に対して超線形に重く、メインスレッドだと大きなグラフの展開で UI がフリーズする。
// elk-worker は Vite が別チャンクに分離するのでメインバンドルにも乗らない。
let elkInstance: ELK | null = null
let elkWorkerFactory: ((url?: string) => Worker) | undefined
function getElk(): ELK {
  if (!elkInstance) {
    elkInstance = new ELKConstructor({ workerUrl: elkWorkerUrl, workerFactory: elkWorkerFactory })
  }
  return elkInstance
}

/** ELK レイアウトで使う worker スクリプトの URL。 */
export { elkWorkerUrl }

/**
 * ELK の Worker の生成方法を差し替える。最初のレイアウトより前に呼ぶ。
 * VS Code の webview は拡張フォルダの URL から直接 Worker を作れない（blob: / data: のみ可）ため、
 * webview 側で worker スクリプトを Blob URL 化した Worker を渡す。
 */
export function setElkWorkerFactory(factory: () => Worker): void {
  elkWorkerFactory = factory
}

type InputNode = {
  id: string
  kind: NodeKind
  label: string
  package: string
  file: string
  line: number
  changed: boolean
  diffStatus: DiffStatus
}

export type LayoutResult = { nodes: AnyFlowNode[]; edges: Edge[] }

/** interface メソッドノードの表示に使う状態（実装の数と展開中の interface メソッド）。 */
export type InterfaceLayoutState = {
  implCounts: Map<string, number>
  expanded: Set<string>
}

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

// 集約エッジの種類: 呼び出しが 1 本でも含まれれば call（実線）にする。
function mergeEdgeKind(a: EdgeKind, b: EdgeKind): EdgeKind {
  return a === 'call' || b === 'call' ? 'call' : 'implements'
}

// implements エッジは diff ステータスの色のまま点線にする。
function edgeStyle(status: DiffStatus, kind: EdgeKind): CSSProperties {
  const style = edgeStyleByStatus(status)
  return kind === 'implements' ? { ...style, strokeDasharray: IMPLEMENTS_EDGE_DASH } : style
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
 * diff ステータスのロジックはレイアウト非依存で、位置のみ ELK に委ねる。
 * クラスタ枠を compound ノードとして扱い、枠を跨ぐ呼び出し連鎖も上→下で読める。
 */
export async function layoutGraph(
  data: GraphResponse,
  expandedClusters: Set<string>,
  interfaces: InterfaceLayoutState,
): Promise<LayoutResult> {
  const inputNodes: InputNode[] = data.graph.nodes.map((n) => ({
    id: n.id,
    kind: n.kind,
    label: nodeDisplayName(n),
    package: n.package,
    file: n.file,
    line: n.line,
    changed: n.changed,
    diffStatus: n.diff_status,
  }))
  const inputEdges: GraphEdge[] = data.graph.edges
  const clusters: Cluster[] = data.clusters

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
  const edgeKindMap = new Map<string, EdgeKind>()
  const edgeKeyToFromTo = new Map<string, { source: string; target: string }>()

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
    const prevKind = edgeKindMap.get(edgeKey)
    edgeKindMap.set(edgeKey, prevKind ? mergeEdgeKind(prevKind, e.kind) : e.kind)
    if (!edgeKeyToFromTo.has(edgeKey)) {
      edgeKeyToFromTo.set(edgeKey, { source: srcId, target: dstId })
    }
  }

  const flowEdges: Edge[] = []
  for (const [edgeKey, status] of edgeStatusMap.entries()) {
    const ft = edgeKeyToFromTo.get(edgeKey)!
    const kind = edgeKindMap.get(edgeKey)!
    flowEdges.push({
      id: `e:${edgeKey}`,
      source: ft.source,
      target: ft.target,
      style: edgeStyle(status, kind),
      markerEnd: edgeMarkerByStatus(status),
      data: { diffStatus: status, kind },
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
            kind: n.kind,
            label: n.label,
            packagePath: n.package,
            file: n.file,
            line: n.line,
            changed: n.changed,
            diffStatus: n.diffStatus,
            clusterId: cid,
            clusterColorHex: color.hex,
            layer,
            implCount: interfaces.implCounts.get(n.id) ?? 0,
            implExpanded: interfaces.expanded.has(n.id),
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

  const laidOut = await getElk().layout(elkGraph)

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
