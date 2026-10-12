import type { Node } from '@xyflow/react'
import type { DiffStatus, NodeKind } from './api'

export type LayerKind = 'ui' | 'domain' | 'data' | 'infra' | 'other'

export type FunctionNodeData = {
  kind: NodeKind
  label: string
  packagePath: string
  file: string
  line: number
  changed: boolean
  diffStatus: DiffStatus
  clusterId: number
  clusterColorHex: string
  layer: LayerKind
  /** interface メソッドの実装の数（関数ノードは 0） */
  implCount: number
  /** interface メソッドの実装を展開中か */
  implExpanded: boolean
  [key: string]: unknown
}

export type ClusterGroupData = {
  label: string
  clusterId: number
  clusterKey: string
  clusterColorHex: string
  clusterColorSoft: string
  [key: string]: unknown
}

export type SuperClusterNodeData = {
  label: string
  clusterId: number
  clusterKey: string
  clusterColorHex: string
  clusterColorSoft: string
  functionCount: number
  changedCount: number
  addedCount: number
  removedCount: number
  hasChanged: boolean
  [key: string]: unknown
}

export type FunctionFlowNode = Node<FunctionNodeData, 'function'>
export type ClusterFlowNode = Node<ClusterGroupData, 'cluster'>
export type SuperClusterFlowNode = Node<SuperClusterNodeData, 'supercluster'>

export type AnyFlowNode = FunctionFlowNode | ClusterFlowNode | SuperClusterFlowNode
