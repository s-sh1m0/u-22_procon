import type { Node } from '@xyflow/react'

export type NodeKind = 'function' | 'file'
export type LayerKind = 'ui' | 'domain' | 'data' | 'infra' | 'other'

export type FunctionNodeData = {
  label: string
  packagePath: string
  file: string
  line: number
  changed: boolean
  clusterId: number
  clusterColorHex: string
  layer: LayerKind
  [key: string]: unknown
}

export type FileNodeData = {
  fileName: string
  packagePath: string
  functionCount: number
  changedCount: number
  changed: boolean
  clusterId: number
  clusterColorHex: string
  layer: LayerKind
  [key: string]: unknown
}

export type ClusterGroupData = {
  label: string
  clusterId: number
  clusterColorHex: string
  clusterColorSoft: string
  [key: string]: unknown
}

export type FunctionFlowNode = Node<FunctionNodeData, 'function'>
export type FileFlowNode = Node<FileNodeData, 'file'>
export type ClusterFlowNode = Node<ClusterGroupData, 'cluster'>

export type AnyFlowNode = FunctionFlowNode | FileFlowNode | ClusterFlowNode
