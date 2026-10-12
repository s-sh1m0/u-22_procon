// JobPhase は running 中のジョブの現在フェーズ（バックエンドの実処理と連動）。
// 配列順が解析パイプラインの進行順と一致する。
export type JobPhase = 'clone' | 'build_graph' | 'diff' | 'cluster' | 'visualize'

export type JobResponse = {
  job_id: string
  status: 'pending' | 'running' | 'done' | 'error'
  phase?: JobPhase
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
  base_sha: string
  head_sha: string
}

export type DiffStatus = 'added' | 'removed' | 'existing'

// NodeKind はノードの種類。interface_method は interface 型のメソッド（本体を持たない）で、
// interface 経由の呼び出しはこのノードを経由して実装メソッドにつながる。
export type NodeKind = 'func' | 'interface_method'

// EdgeKind はエッジの種類。implements は interface メソッドから実装メソッドへのエッジ。
export type EdgeKind = 'call' | 'implements'

export type ClusterMode = 'louvain' | 'package' | 'file'

export type GraphResponse = {
  pr: PRInfo
  clusters: Cluster[]
  graph: Graph
  cycles: Cycle[]
  violations: LayerViolation[]
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
  kind: NodeKind
  name: string
  // メソッドのレシーバ型名（ポインタ記号なし）。関数は空文字。
  receiver: string
  package: string
  file: string
  line: number
  changed: boolean
  diff_status: DiffStatus
}

export type GraphEdge = {
  from: string
  to: string
  kind: EdgeKind
  status: DiffStatus
}

export type Cycle = {
  id: number
  nodes: string[]
  is_new: boolean
}

// LayerViolation は依存方向の逆転（内側レイヤー → 外側レイヤー）を表す。
// 例: from_layer="domain" / to_layer="infra" は DIP に反するアンチパターン。
export type LayerViolation = {
  from: string
  to: string
  from_layer: string
  to_layer: string
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
