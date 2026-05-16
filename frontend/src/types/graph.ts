import type { Node } from '@xyflow/react'
import type { DiffStatus } from './api'

export type NodeKind = 'function' | 'file'
export type LayerKind = 'ui' | 'domain' | 'data' | 'infra' | 'other'

export type FunctionNodeData = {
  label: string
  packagePath: string
  file: string
  line: number
  changed: boolean
  diffStatus: DiffStatus
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
  addedCount: number
  removedCount: number
  changed: boolean
  diffStatus: DiffStatus
  clusterId: number
  clusterColorHex: string
  layer: LayerKind
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
export type FileFlowNode = Node<FileNodeData, 'file'>
export type ClusterFlowNode = Node<ClusterGroupData, 'cluster'>
export type SuperClusterFlowNode = Node<SuperClusterNodeData, 'supercluster'>

export type AnyFlowNode = FunctionFlowNode | FileFlowNode | ClusterFlowNode | SuperClusterFlowNode
