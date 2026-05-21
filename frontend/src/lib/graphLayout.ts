import type { CSSProperties } from 'react'
import { MarkerType, type Edge, type EdgeMarker } from '@xyflow/react'
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

const LAYER_ORDER: LayerKind[] = ['ui', 'domain', 'data', 'infra', 'other']

const NODE_W = 200
const NODE_H = 60
const COLS = 5
const COL_STRIDE = 220
const ROW_STRIDE = 140
const PADDING = 24
const CLUSTER_LABEL_H = 28
const CLUSTER_GAP = 48
const LAYER_GAP = 64

const SUPER_W = 240
const SUPER_H = 90

type InputNode = {
  id: string
  name: string
  package: string
  file: string
  line: number
  changed: boolean
  diffStatus: DiffStatus
  functionCount: number
  changedCount: number
  addedCount: number
  removedCount: number
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

export function layoutGraph(data: GraphResponse, expandedClusters: Set<string>): LayoutResult {
  const inputNodes: InputNode[] = data.graph.nodes.map((n) => ({
    id: n.id,
    name: n.name,
    package: n.package,
    file: n.file,
    line: n.line,
    changed: n.changed,
    diffStatus: n.diff_status,
    functionCount: 1,
    changedCount: n.changed ? 1 : 0,
    addedCount: n.diff_status === 'added' ? 1 : 0,
    removedCount: n.diff_status === 'removed' ? 1 : 0,
  }))
  const inputEdges: GraphEdge[] = data.graph.edges
  const clusters: Cluster[] = data.clusters

  const nodeToCluster = new Map<string, number>()
  for (const c of clusters) {
    for (const nid of c.nodes) nodeToCluster.set(nid, c.id)
  }

  // Build a quick lookup for nodes
  const nodeMap = new Map<string, InputNode>()
  for (const n of inputNodes) nodeMap.set(n.id, n)

  // Group nodes by (layer, clusterId)
  const groups = new Map<string, InputNode[]>()
  for (const n of inputNodes) {
    const layer = inferLayer(n.package)
    const cid = nodeToCluster.get(n.id) ?? 0
    const key = `${layer}:${cid}`
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(n)
  }

  const flowNodes: AnyFlowNode[] = []

  // Build edges with cluster aggregation. DiffStatus は added > removed > existing で昇格。
  const edgeStatusMap = new Map<string, DiffStatus>()
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
    if (!edgeKeyToFromTo.has(edgeKey)) {
      edgeKeyToFromTo.set(edgeKey, { source: srcId, target: dstId })
    }
  }

  const flowEdges: Edge[] = []
  for (const [edgeKey, status] of edgeStatusMap.entries()) {
    const ft = edgeKeyToFromTo.get(edgeKey)!
    flowEdges.push({
      id: `e:${edgeKey}`,
      source: ft.source,
      target: ft.target,
      style: edgeStyleByStatus(status),
      markerEnd: edgeMarkerByStatus(status),
      data: { diffStatus: status },
    })
  }

  let layerY = 0

  for (const layer of LAYER_ORDER) {
    const layerGroups: Array<{ cid: number; nodes: InputNode[] }> = []
    for (const [key, nodes] of groups.entries()) {
      const colonIdx = key.indexOf(':')
      const l = key.slice(0, colonIdx)
      const cidStr = key.slice(colonIdx + 1)
      if (l === layer) {
        layerGroups.push({ cid: Number(cidStr), nodes })
      }
    }
    if (layerGroups.length === 0) continue

    let clusterX = 0
    let maxClusterH = 0

    for (const { cid, nodes } of layerGroups) {
      const color = getClusterColor(cid)
      const clusterLabel = clusters.find((c) => c.id === cid)?.label ?? `Cluster ${cid}`
      const key = `${layer}:${cid}`
      const isExpanded = expandedClusters.has(key)

      if (!isExpanded) {
        // Collapsed: render a single super node
        const changedCount = nodes.reduce((acc, n) => acc + n.changedCount, 0)
        const functionCount = nodes.reduce((acc, n) => acc + n.functionCount, 0)
        const addedCount = nodes.reduce((acc, n) => acc + n.addedCount, 0)
        const removedCount = nodes.reduce((acc, n) => acc + n.removedCount, 0)

        flowNodes.push({
          id: `super:${key}`,
          type: 'supercluster',
          position: { x: clusterX, y: layerY },
          data: {
            label: clusterLabel,
            clusterId: cid,
            clusterKey: key,
            clusterColorHex: color.hex,
            clusterColorSoft: color.soft,
            functionCount,
            changedCount,
            addedCount,
            removedCount,
            hasChanged: changedCount > 0 || addedCount > 0 || removedCount > 0,
          } as SuperClusterNodeData,
          style: { width: SUPER_W, height: SUPER_H },
        })

        clusterX += SUPER_W + CLUSTER_GAP
        maxClusterH = Math.max(maxClusterH, SUPER_H)
      } else {
        // Expanded: render cluster container + child nodes
        const rows = Math.ceil(nodes.length / COLS)
        const actualCols = Math.min(nodes.length, COLS)
        const clusterW = PADDING * 2 + (actualCols - 1) * COL_STRIDE + NODE_W
        const clusterH =
          PADDING * 2 + CLUSTER_LABEL_H + rows * NODE_H + (rows - 1) * (ROW_STRIDE - NODE_H)

        const clusterId = `cluster:${layer}:${cid}`

        flowNodes.push({
          id: clusterId,
          type: 'cluster',
          position: { x: clusterX, y: layerY },
          data: {
            label: clusterLabel,
            clusterId: cid,
            clusterKey: key,
            clusterColorHex: color.hex,
            clusterColorSoft: color.soft,
          } as ClusterGroupData,
          style: { width: clusterW, height: clusterH },
          draggable: false,
          selectable: true,
          zIndex: 0,
        })

        nodes.forEach((n, i) => {
          const col = i % COLS
          const row = Math.floor(i / COLS)
          const x = PADDING + col * COL_STRIDE
          const y = PADDING + CLUSTER_LABEL_H + row * ROW_STRIDE

          flowNodes.push({
            id: n.id,
            type: 'function',
            parentId: clusterId,
            extent: 'parent',
            position: { x, y },
            zIndex: 1,
            data: {
              label: n.name,
              packagePath: n.package,
              file: n.file,
              line: n.line,
              changed: n.changed,
              diffStatus: n.diffStatus,
              clusterId: cid,
              clusterColorHex: color.hex,
              layer,
            } as FunctionNodeData,
          })
        })

        clusterX += clusterW + CLUSTER_GAP
        maxClusterH = Math.max(maxClusterH, clusterH)
      }
    }

    layerY += maxClusterH + LAYER_GAP
  }

  return { nodes: flowNodes, edges: flowEdges }
}
