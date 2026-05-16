export type JobResponse = {
  job_id: string
  status: 'pending' | 'running' | 'done' | 'error'
  analysis_id?: string
  error?: string
}

export type PRInfo = {
  owner: string
  repo: string
  number: number
  title: string
  base_ref: string
  head_ref: string
}

export type DiffStatus = 'added' | 'removed' | 'existing'

export type ClusterMode = 'louvain' | 'package' | 'file'

export type GraphResponse = {
  pr: PRInfo
  clusters: Cluster[]
  graph: Graph
  cycles: Cycle[]
}

export type Cluster = {
  id: number
  label: string
  nodes: string[]
}

export type Graph = {
  nodes: GraphNode[]
  edges: GraphEdge[]
}

export type GraphNode = {
  id: string
  name: string
  package: string
  file: string
  line: number
  changed: boolean
  diff_status: DiffStatus
}

export type GraphEdge = {
  from: string
  to: string
  status: DiffStatus
}

export type Cycle = {
  id: number
  nodes: string[]
  is_new: boolean
}

export type DiffFile = {
  filename: string
  previous_name?: string
  status: 'added' | 'modified' | 'removed' | 'renamed'
  additions: number
  deletions: number
  before_content: string
  after_content: string
}

export type DiffResponse = {
  files: DiffFile[]
}
