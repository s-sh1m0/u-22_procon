import { useMemo, useCallback, useEffect, useRef, useState } from 'react'
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
import GraphSearch, { type SearchCandidate } from './GraphSearch'
import FocusLegend from './FocusLegend'
import { layoutGraph, type LayoutResult } from '@/lib/graphLayout'
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
  searchCandidates: SearchCandidate[]
  onJump: (id: string) => void
  centerTarget: { nodeId: string; nonce: number } | null
  changedCount: number
  changedIndex: number | null
  onNextChanged: () => void
  onPrevChanged: () => void
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
  searchCandidates,
  onJump,
  centerTarget,
  changedCount,
  changedIndex,
  onNextChanged,
  onPrevChanged,
}: Props) {
  const { fitView } = useReactFlow()

  // ELK レイアウトは非同期。結果を state に持ち、どの data から算出したかを source で保持する
  // （新しいグラフに対してだけ fitView するため）。再計算中は前回のレイアウトを表示し続ける。
  // レイアウト結果と、それを算出した入力（data / expandedClusters）を保持する。
  // 現在の入力と食い違っている間 = 再計算中、として isLayouting を導出する
  // （effect 内で同期 setState せず派生値で表すため）。
  const [layout, setLayout] = useState<
    LayoutResult & { source: GraphResponse | null; expanded: Set<string> | null }
  >({ nodes: [], edges: [], source: null, expanded: null })

  useEffect(() => {
    let cancelled = false
    layoutGraph(data, expandedClusters)
      .then((res) => {
        if (!cancelled) setLayout({ ...res, source: data, expanded: expandedClusters })
      })
      .catch((err) => {
        console.error('graph layout failed', err)
      })
    return () => {
      cancelled = true
    }
  }, [data, expandedClusters])

  const { nodes: layoutNodes, edges: layoutEdges, source: layoutSource } = layout
  const isLayouting = layout.source !== data || layout.expanded !== expandedClusters

  // 新しいグラフ（data 変更）のレイアウトが整ったら一度だけ全体にフィットする。
  // 展開/折りたたみ（expandedClusters 変更）では data 不変なので再フィットせず位置を保つ。
  const fittedSourceRef = useRef<GraphResponse | null>(null)
  useEffect(() => {
    if (layoutSource !== data || layoutNodes.length === 0) return
    if (fittedSourceRef.current === data) return
    fittedSourceRef.current = data
    const raf = requestAnimationFrame(() => fitView({ duration: 300 }))
    return () => cancelAnimationFrame(raf)
  }, [layoutSource, data, layoutNodes, fitView])

  // 検索ジャンプの中央寄せ。ジャンプ要求（centerTarget.nonce 変化）を pending に記録し、
  // 対象ノードが現在のレイアウトに現れ次第そこへ寄せる。折りたたみクラスタ内のノードは
  // 親展開 → 再レイアウト後に初めて layoutNodes に現れるため、layoutNodes 変化でも再評価する。
  const pendingCenterRef = useRef<string | null>(null)
  const processedNonceRef = useRef<number>(-1)
  useEffect(() => {
    if (centerTarget && centerTarget.nonce !== processedNonceRef.current) {
      processedNonceRef.current = centerTarget.nonce
      pendingCenterRef.current = centerTarget.nodeId
    }
    const targetId = pendingCenterRef.current
    if (!targetId) return
    if (!layoutNodes.some((n) => n.id === targetId)) return // まだ配置されていない（展開待ち）
    pendingCenterRef.current = null
    const raf = requestAnimationFrame(() =>
      fitView({ nodes: [{ id: targetId }], duration: 500, maxZoom: 1.2, padding: 0.6 }),
    )
    return () => cancelAnimationFrame(raf)
  }, [centerTarget, layoutNodes, fitView])

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
      {isLayouting && (
        <Panel position="top-center">
          <div className="flex items-center gap-2 rounded-full border border-stone-200 bg-white/90 px-3 py-1.5 text-xs text-stone-500 shadow-sm backdrop-blur">
            <svg
              width="14"
              height="14"
              viewBox="0 0 20 20"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              className="animate-spin"
            >
              <circle cx="10" cy="10" r="7" strokeDasharray="32" strokeDashoffset="8" />
            </svg>
            レイアウト計算中…
          </div>
        </Panel>
      )}
      <Panel position="top-left" className="flex flex-col gap-2">
        <GraphSearch candidates={searchCandidates} onJump={onJump} />
        {focus && <FocusLegend callerCount={focus.callerCount} calleeCount={focus.calleeCount} />}
      </Panel>
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
          changedCount={changedCount}
          changedIndex={changedIndex}
          onNextChanged={onNextChanged}
          onPrevChanged={onPrevChanged}
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
