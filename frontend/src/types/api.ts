// TODO: バックエンド API レスポンスの TypeScript 型定義

export interface JobResponse {
  job_id: string;
  status: "pending" | "running" | "done" | "error";
  analysis_id?: string;
  error?: string;
}

export interface AnalysisResponse {
  id: string;
  pr: PRInfo;
  result: ClusterResult;
  created_at: string;
}

export interface PRInfo {
  owner: string;
  repo: string;
  number: number;
  title: string;
}

export interface ClusterResult {
  clusters: Cluster[];
  graph: Graph;
}

export interface Cluster {
  id: number;
  label: string;
  node_ids: string[];
}

export interface Graph {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface GraphNode {
  id: string;
  name: string;
  package: string;
  file: string;
  line: number;
  changed: boolean;
}

export interface GraphEdge {
  from: string;
  to: string;
}
