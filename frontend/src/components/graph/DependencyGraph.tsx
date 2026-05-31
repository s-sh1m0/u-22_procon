import { useMemo, useCallback } from 'react'
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  ReactFlowProvider,
  Panel,
  useReactFlow,
  type NodeMouseHandler,
} from '@xyflow/react'
import type { ClusterMode, GraphResponse } from '@/types/api'
import FunctionNode from './FunctionNode'
import ClusterGroup from './ClusterGroup'
import SuperClusterNode from './SuperClusterNode'
import GraphControls from './GraphControls'
import FocusLegend from './FocusLegend'
import { layoutGraph } from '@/lib/graphLayout'
import {
  computeFocus,
  focusEdgeStyle,
  focusEdgeMarker,
  DIMMED_EDGE_STYLE,
  FOCUSED_OPACITY,
  DIMMED_OPACITY,
} from '@/lib/graphFocus'

const nodeTypes = {
  function: FunctionNode,
  cluster: ClusterGroup,
  supercluster: SuperClusterNode,
}

// 描画ノード数がこれを超えたら MiniMap を省略する。MiniMap はノード1個につき
// SVG 矩形を1個描くため、大規模グラフでは常時コストになる。
const MINIMAP_HIDE_NODE_THRESHOLD = 150

// フォーカス時にエッジアニメーション（流れる破線）を行う最大本数。
// ハブノード選択で数百本が同時にアニメーションすると重くなるため上限を設ける。
const FOCUS_ANIMATE_MAX_EDGES = 60

type Props = {
  data: GraphResponse
  clusterMode: ClusterMode
  onChangeClusterMode: (mode: ClusterMode) => void
  impactOnly: boolean
  onSetImpactOnly: (next: boolean) => void
  selectedNodeId: string | null
  onSelectNode: (id: string) => void
  onClearSelection: () => void
  expandedClusters: Set<string>
  onToggleCluster: (key: string) => void
  onExpandAll: () => void
  onCollapseAll: () => void
}

function GraphInner({
  data,
  clusterMode,
  onChangeClusterMode,
  impactOnly,
  onSetImpactOnly,
  selectedNodeId,
  onSelectNode,
  onClearSelection,
  expandedClusters,
  onToggleCluster,
  onExpandAll,
  onCollapseAll,
}: Props) {
  const { fitView } = useReactFlow()

  const { nodes: layoutNodes, edges: layoutEdges } = useMemo(
    () => layoutGraph(data, expandedClusters),
    [data, expandedClusters],
  )

  // フォーカス: 選択ノードが現在のレイアウトに存在するときだけ有効化する
  // （クラスタ折りたたみ等で非表示になった選択は無視し、全体減光を避ける）。
  const focus = useMemo(() => {
    if (!selectedNodeId) return null
    if (!layoutNodes.some((n) => n.id === selectedNodeId)) return null
    return computeFocus(layoutEdges, selectedNodeId)
  }, [selectedNodeId, layoutNodes, layoutEdges])

  // フォーカス中に通常表示するクラスタコンテナ = フォーカス対象の関数ノードの親
  const focusedContainerIds = useMemo(() => {
    if (!focus) return null
    const ids = new Set<string>()
    for (const n of layoutNodes) {
      if (n.type === 'function' && n.parentId && focus.nodeIds.has(n.id)) ids.add(n.parentId)
    }
    return ids
  }, [focus, layoutNodes])

  const nodes = useMemo(
    () =>
      layoutNodes.map((n) => {
        const base = { ...n, selected: n.id === selectedNodeId }
        if (!focus) return base
        const inFocus =
          n.type === 'cluster' ? (focusedContainerIds?.has(n.id) ?? false) : focus.nodeIds.has(n.id)
        return {
          ...base,
          style: { ...n.style, opacity: inFocus ? FOCUSED_OPACITY : DIMMED_OPACITY },
        }
      }),
    [layoutNodes, selectedNodeId, focus, focusedContainerIds],
  )

  const edges = useMemo(() => {
    if (!focus) return layoutEdges
    // フォーカスエッジが多すぎるときはアニメーションを切る（描画負荷対策）。
    const animate = focus.edgeDir.size <= FOCUS_ANIMATE_MAX_EDGES
    return layoutEdges.map((e) => {
      const dir = focus.edgeDir.get(e.id)
      if (!dir) return { ...e, style: DIMMED_EDGE_STYLE, markerEnd: undefined, animated: false }
      return {
        ...e,
        style: focusEdgeStyle(dir),
        markerEnd: focusEdgeMarker(dir),
        animated: animate,
      }
    })
  }, [layoutEdges, focus])

  const handleNodeClick: NodeMouseHandler = useCallback(
    (_, node) => {
      if (node.type === 'supercluster' || node.type === 'cluster') {
        onToggleCluster(node.data.clusterKey as string)
      } else if (node.type === 'function') {
        onSelectNode(node.id)
      }
    },
    [onSelectNode, onToggleCluster],
  )

  const handleFitChanged = useCallback(() => {
    const changedNodes = nodes.filter((n) => {
      if (n.type !== 'function') return false
      const d = n.data as { changed?: boolean; diffStatus?: string }
      return d.changed || (d.diffStatus && d.diffStatus !== 'existing')
    })
    if (changedNodes.length > 0) {
      fitView({ nodes: changedNodes.map((n) => ({ id: n.id })), duration: 400, padding: 0.3 })
    } else {
      fitView({ duration: 300 })
    }
  }, [fitView, nodes])

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      fitView
      minZoom={0.05}
      maxZoom={2.5}
      onlyRenderVisibleElements
      nodesDraggable={false}
      nodesConnectable={false}
      elevateNodesOnSelect={false}
      onNodeClick={handleNodeClick}
      onPaneClick={onClearSelection}
      proOptions={{ hideAttribution: false }}
    >
      <Background variant={BackgroundVariant.Dots} gap={20} color="#e7e5e4" />
      {focus && (
        <Panel position="top-left">
          <FocusLegend callerCount={focus.callerCount} calleeCount={focus.calleeCount} />
        </Panel>
      )}
      <Controls position="bottom-left" />
      {nodes.length <= MINIMAP_HIDE_NODE_THRESHOLD && (
        <MiniMap
          nodeColor={(n) =>
            n.type === 'cluster' || n.type === 'supercluster' ? '#e7e5e4' : '#0d9488'
          }
          maskColor="rgba(250,250,249,0.6)"
        />
      )}
      <Panel position="bottom-right">
        <GraphControls
          clusterMode={clusterMode}
          onChangeClusterMode={onChangeClusterMode}
          impactOnly={impactOnly}
          onSetImpactOnly={onSetImpactOnly}
          onFitChanged={handleFitChanged}
          onExpandAll={onExpandAll}
          onCollapseAll={onCollapseAll}
        />
      </Panel>
    </ReactFlow>
  )
}

export default function DependencyGraph(props: Props) {
  return (
    <ReactFlowProvider>
      <div className="size-full min-h-0">
        <GraphInner {...props} />
      </div>
    </ReactFlowProvider>
  )
}
