import { useMemo, useCallback, useEffect } from 'react'
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
import type { GraphResponse } from '@/types/api'
import type { NodeKind } from '@/types/graph'
import FunctionNode from './FunctionNode'
import FileNode from './FileNode'
import ClusterGroup from './ClusterGroup'
import SuperClusterNode from './SuperClusterNode'
import GraphControls from './GraphControls'
import { layoutGraph } from '@/lib/graphLayout'

const nodeTypes = {
  function: FunctionNode,
  file: FileNode,
  cluster: ClusterGroup,
  supercluster: SuperClusterNode,
}

type Props = {
  data: GraphResponse
  nodeKind: NodeKind
  onChangeNodeKind: (kind: NodeKind) => void
  selectedNodeId: string | null
  onSelectNode: (id: string) => void
  expandedClusters: Set<string>
  onToggleCluster: (key: string) => void
  onExpandAll: () => void
  onCollapseAll: () => void
  /** 値が変わるたびに該当ノードへ fitView する。CycleAlert からのフォーカス用。 */
  focusNodeId?: string | null
}

function GraphInner({
  data,
  nodeKind,
  onChangeNodeKind,
  selectedNodeId,
  onSelectNode,
  expandedClusters,
  onToggleCluster,
  onExpandAll,
  onCollapseAll,
  focusNodeId,
}: Props) {
  const { fitView } = useReactFlow()

  const { nodes: layoutNodes, edges } = useMemo(
    () => layoutGraph(data, nodeKind, expandedClusters),
    [data, nodeKind, expandedClusters],
  )

  const nodes = useMemo(
    () => layoutNodes.map((n) => ({ ...n, selected: n.id === selectedNodeId })),
    [layoutNodes, selectedNodeId],
  )

  const handleNodeClick: NodeMouseHandler = useCallback(
    (_, node) => {
      if (node.type === 'supercluster' || node.type === 'cluster') {
        onToggleCluster(node.data.clusterKey as string)
      } else if (node.type === 'function' || node.type === 'file') {
        onSelectNode(node.id)
      }
    },
    [onSelectNode, onToggleCluster],
  )

  const handleFitChanged = useCallback(() => {
    const changedNodes = nodes.filter((n) => {
      if (n.type !== 'function' && n.type !== 'file') return false
      const d = n.data as { changed?: boolean; diffStatus?: string }
      return d.changed || (d.diffStatus && d.diffStatus !== 'existing')
    })
    if (changedNodes.length > 0) {
      fitView({ nodes: changedNodes.map((n) => ({ id: n.id })), duration: 400, padding: 0.3 })
    } else {
      fitView({ duration: 300 })
    }
  }, [fitView, nodes])

  // focusNodeId が変わったら該当ノードに fitView。
  // 折りたたみ中のクラスタに属するときはスーパーノード（または親クラスタコンテナ）にフォーカスする。
  useEffect(() => {
    if (!focusNodeId) return
    const target = nodes.find((n) => n.id === focusNodeId)
    if (target) {
      fitView({ nodes: [{ id: target.id }], duration: 400, padding: 0.4 })
      return
    }
    // 該当ノードがレンダリングされていない場合（クラスタ折りたたみ中など）は親またはスーパーノードを探す
    const superMatch = nodes.find(
      (n) => n.type === 'supercluster' && data.clusters.some((c) => c.nodes.includes(focusNodeId)),
    )
    if (superMatch) {
      fitView({ nodes: [{ id: superMatch.id }], duration: 400, padding: 0.4 })
    }
  }, [focusNodeId, nodes, fitView, data.clusters])

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      nodeTypes={nodeTypes}
      fitView
      minZoom={0.05}
      maxZoom={2.5}
      onNodeClick={handleNodeClick}
      proOptions={{ hideAttribution: false }}
    >
      <Background variant={BackgroundVariant.Dots} gap={20} color="#e7e5e4" />
      <Controls position="bottom-left" />
      <MiniMap
        nodeColor={(n) =>
          n.type === 'cluster' || n.type === 'supercluster' ? '#e7e5e4' : '#0d9488'
        }
        maskColor="rgba(250,250,249,0.6)"
      />
      <Panel position="bottom-right">
        <GraphControls
          nodeKind={nodeKind}
          onChangeNodeKind={onChangeNodeKind}
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
