import type { GraphResponse, GraphEdge, Cluster, DiffStatus } from '@/types/api'

export type AggregatedFileNode = {
  id: string
  name: string
  package: string
  file: string
  line: number
  changed: boolean
  diffStatus: DiffStatus
  functionCount: number
  changedCount: number
  addedCount: number
  removedCount: number
}

export type FileAggResult = {
  nodes: AggregatedFileNode[]
  edges: GraphEdge[]
  clusters: Cluster[]
}

// 複数エッジの DiffStatus を集約。優先順: added > removed > existing
function mergeEdgeStatus(a: DiffStatus, b: DiffStatus): DiffStatus {
  if (a === 'added' || b === 'added') return 'added'
  if (a === 'removed' || b === 'removed') return 'removed'
  return 'existing'
}

// ファイル単位の DiffStatus を決定する。全 added/全 removed のときのみ集約、それ以外は existing。
function deriveFileStatus(addedCount: number, removedCount: number, total: number): DiffStatus {
  if (total === 0) return 'existing'
  if (addedCount === total) return 'added'
  if (removedCount === total) return 'removed'
  return 'existing'
}

export function aggregateByFile(data: GraphResponse): FileAggResult {
  const { graph, clusters } = data

  const nodeToCluster = new Map<string, number>()
  for (const c of clusters) {
    for (const nid of c.nodes) nodeToCluster.set(nid, c.id)
  }

  // Group function nodes by file path
  const fileGroups = new Map<string, typeof graph.nodes>()
  for (const n of graph.nodes) {
    const fileKey = n.file || n.package
    const arr = fileGroups.get(fileKey)
    if (arr) arr.push(n)
    else fileGroups.set(fileKey, [n])
  }

  const nodeToFileId = new Map<string, string>()
  const fileNodes: AggregatedFileNode[] = []

  for (const [file, nodes] of fileGroups.entries()) {
    const fileNodeId = `file:${file}`
    for (const n of nodes) nodeToFileId.set(n.id, fileNodeId)

    const changedCount = nodes.filter((n) => n.changed).length
    const addedCount = nodes.filter((n) => n.diff_status === 'added').length
    const removedCount = nodes.filter((n) => n.diff_status === 'removed').length
    const pkg = nodes[0]?.package ?? ''
    const fileName = file.split('/').pop() ?? file

    fileNodes.push({
      id: fileNodeId,
      name: fileName,
      package: pkg,
      file,
      line: 0,
      changed: changedCount > 0,
      diffStatus: deriveFileStatus(addedCount, removedCount, nodes.length),
      functionCount: nodes.length,
      changedCount,
      addedCount,
      removedCount,
    })
  }

  // Deduplicate inter-file edges, merging DiffStatus by priority added > removed > existing
  const edgeStatus = new Map<string, DiffStatus>()
  const edgeKeyToFromTo = new Map<string, { from: string; to: string }>()
  for (const e of graph.edges) {
    const fromFile = nodeToFileId.get(e.from)
    const toFile = nodeToFileId.get(e.to)
    if (!fromFile || !toFile || fromFile === toFile) continue
    const key = `${fromFile}->${toFile}`
    const prev = edgeStatus.get(key)
    edgeStatus.set(key, prev ? mergeEdgeStatus(prev, e.status) : e.status)
    if (!edgeKeyToFromTo.has(key)) edgeKeyToFromTo.set(key, { from: fromFile, to: toFile })
  }
  const fileEdges: GraphEdge[] = []
  for (const [key, st] of edgeStatus.entries()) {
    const ft = edgeKeyToFromTo.get(key)!
    fileEdges.push({ from: ft.from, to: ft.to, status: st })
  }

  // Assign each file node to the dominant cluster of its functions
  const fileNodeClusterFreq = new Map<string, Map<number, number>>()
  for (const n of graph.nodes) {
    const fileNodeId = nodeToFileId.get(n.id)
    if (!fileNodeId) continue
    const cid = nodeToCluster.get(n.id) ?? 0
    if (!fileNodeClusterFreq.has(fileNodeId)) fileNodeClusterFreq.set(fileNodeId, new Map())
    const freq = fileNodeClusterFreq.get(fileNodeId)!
    freq.set(cid, (freq.get(cid) ?? 0) + 1)
  }

  const newClusterMembers = new Map<number, string[]>()
  for (const [fileNodeId, freq] of fileNodeClusterFreq.entries()) {
    let dominantCid = 0
    let maxCount = 0
    for (const [cid, count] of freq.entries()) {
      if (count > maxCount) {
        maxCount = count
        dominantCid = cid
      }
    }
    const arr = newClusterMembers.get(dominantCid)
    if (arr) arr.push(fileNodeId)
    else newClusterMembers.set(dominantCid, [fileNodeId])
  }

  const fileClusters: Cluster[] = []
  for (const origCluster of clusters) {
    const members = newClusterMembers.get(origCluster.id)
    if (members && members.length > 0) {
      fileClusters.push({ id: origCluster.id, label: origCluster.label, nodes: members })
    }
  }

  return { nodes: fileNodes, edges: fileEdges, clusters: fileClusters }
}
