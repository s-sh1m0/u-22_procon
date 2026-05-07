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
import type { GraphResponse } from '@/types/api'
import type { NodeKind } from '@/types/graph'
import FunctionNode from './FunctionNode'
import FileNode from './FileNode'
import ClusterGroup from './ClusterGroup'
import GraphControls from './GraphControls'
import { layoutGraph } from '@/lib/graphLayout'

const nodeTypes = {
  function: FunctionNode,
  file: FileNode,
  cluster: ClusterGroup,
}

type Props = {
  data: GraphResponse
  nodeKind: NodeKind
  onChangeNodeKind: (kind: NodeKind) => void
  selectedNodeId: string | null
  onSelectNode: (id: string) => void
}

function GraphInner({ data, nodeKind, onChangeNodeKind, selectedNodeId, onSelectNode }: Props) {
  const { fitView } = useReactFlow()

  const { nodes: layoutNodes, edges } = useMemo(() => layoutGraph(data, nodeKind), [data, nodeKind])

  const nodes = useMemo(
    () => layoutNodes.map((n) => ({ ...n, selected: n.id === selectedNodeId })),
    [layoutNodes, selectedNodeId],
  )

  const handleNodeClick: NodeMouseHandler = useCallback(
    (_, node) => {
      if (node.type === 'function' || node.type === 'file') {
        onSelectNode(node.id)
      }
    },
    [onSelectNode],
  )

  const handleFitChanged = useCallback(() => {
    const changedNodes = nodes.filter(
      (n) =>
        (n.type === 'function' || n.type === 'file') && (n.data as { changed?: boolean }).changed,
    )
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
      onNodeClick={handleNodeClick}
      proOptions={{ hideAttribution: false }}
    >
      <Background variant={BackgroundVariant.Dots} gap={20} color="#e7e5e4" />
      <Controls position="bottom-left" />
      <MiniMap
        nodeColor={(n) => (n.type === 'cluster' ? '#e7e5e4' : '#0d9488')}
        maskColor="rgba(250,250,249,0.6)"
      />
      <Panel position="bottom-right">
        <GraphControls
          nodeKind={nodeKind}
          onChangeNodeKind={onChangeNodeKind}
          onFitChanged={handleFitChanged}
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
