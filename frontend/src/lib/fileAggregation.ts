import type { GraphResponse, GraphEdge, Cluster } from '@/types/api'

export type AggregatedFileNode = {
  id: string
  name: string
  package: string
  file: string
  line: number
  changed: boolean
  functionCount: number
  changedCount: number
}

export type FileAggResult = {
  nodes: AggregatedFileNode[]
  edges: GraphEdge[]
  clusters: Cluster[]
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
    const pkg = nodes[0]?.package ?? ''
    const fileName = file.split('/').pop() ?? file

    fileNodes.push({
      id: fileNodeId,
      name: fileName,
      package: pkg,
      file,
      line: 0,
      changed: changedCount > 0,
      functionCount: nodes.length,
      changedCount,
    })
  }

  // Deduplicate inter-file edges
  const edgeSet = new Set<string>()
  const fileEdges: GraphEdge[] = []
  for (const e of graph.edges) {
    const fromFile = nodeToFileId.get(e.from)
    const toFile = nodeToFileId.get(e.to)
    if (!fromFile || !toFile || fromFile === toFile) continue
    const key = `${fromFile}->${toFile}`
    if (edgeSet.has(key)) continue
    edgeSet.add(key)
    fileEdges.push({ from: fromFile, to: toFile })
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
