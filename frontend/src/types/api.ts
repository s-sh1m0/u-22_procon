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

export type GraphResponse = {
  pr: PRInfo
  clusters: Cluster[]
  graph: Graph
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
}

export type GraphEdge = {
  from: string
  to: string
}
