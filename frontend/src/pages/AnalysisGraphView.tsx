import { useMemo, useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import type { NodeKind, LayerKind, AnyFlowNode } from '@/types/graph'
import type { Cluster, DiffFile } from '@/types/api'
import { useGraph } from '@/hooks/useGraph'
import { useDiff } from '@/hooks/useDiff'
import { inferLayer } from '@/lib/layerInference'
import { getClusterColor } from '@/lib/clusterColors'
import AppShell from '@/components/layout/AppShell'
import PRMetaBar from '@/components/layout/PRMetaBar'
import DependencyGraph from '@/components/graph/DependencyGraph'
import FunctionDetailsPanel from '@/components/graph/FunctionDetailsPanel'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

type Props = { jobId: string }

export default function AnalysisGraphView({ jobId }: Props) {
  const { data, isLoading, error } = useGraph(jobId, true)
  const { data: diffData } = useDiff(jobId, !isLoading && !error && !!data)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [nodeKind, setNodeKind] = useState<NodeKind>('function')

  const handleChangeNodeKind = useCallback((kind: NodeKind) => {
    setNodeKind(kind)
    setSelectedNodeId(null)
  }, [])

  const { panelNode, selectedCluster, selectedLayer, panelDiff } = useMemo(() => {
    if (!data || !selectedNodeId) {
      return {
        panelNode: null,
        selectedCluster: null,
        selectedLayer: 'other' as LayerKind,
        panelDiff: undefined as DiffFile | undefined,
      }
    }

    // File node selected (file mode)
    if (selectedNodeId.startsWith('file:')) {
      const filePath = selectedNodeId.slice(5)
      const fileNodes = data.graph.nodes.filter((n) => n.file === filePath)
      if (fileNodes.length === 0) {
        return { panelNode: null, selectedCluster: null, selectedLayer: 'other' as LayerKind }
      }
      const pkg = fileNodes[0].package
      const layer = inferLayer(pkg)
      const changedCount = fileNodes.filter((n) => n.changed).length

      const clusterFreq = new Map<number, number>()
      for (const n of fileNodes) {
        const c = data.clusters.find((c) => c.nodes.includes(n.id))
        if (c) clusterFreq.set(c.id, (clusterFreq.get(c.id) ?? 0) + 1)
      }
      let domCid = 0
      let maxFreq = 0
      for (const [cid, freq] of clusterFreq.entries()) {
        if (freq > maxFreq) {
          maxFreq = freq
          domCid = cid
        }
      }
      const cluster = data.clusters.find((c) => c.id === domCid) ?? null
      const color = getClusterColor(domCid)

      const node: AnyFlowNode = {
        id: selectedNodeId,
        type: 'file',
        position: { x: 0, y: 0 },
        data: {
          fileName: filePath.split('/').pop() ?? filePath,
          packagePath: pkg,
          functionCount: fileNodes.length,
          changedCount,
          changed: changedCount > 0,
          clusterId: domCid,
          clusterColorHex: color.hex,
          layer,
        },
      }
      const fileDiff = diffData?.files.find(
        (f) => filePath === f.filename || filePath.endsWith('/' + f.filename),
      )
      return {
        panelNode: node,
        selectedCluster: cluster,
        selectedLayer: layer,
        panelDiff: fileDiff,
      }
    }

    // Function node selected
    const gn = data.graph.nodes.find((n) => n.id === selectedNodeId)
    if (!gn) return { panelNode: null, selectedCluster: null, selectedLayer: 'other' as LayerKind }

    const layer = inferLayer(gn.package)
    const cluster: Cluster | null = data.clusters.find((c) => c.nodes.includes(gn.id)) ?? null
    const cid = cluster?.id ?? 0
    const color = getClusterColor(cid)

    const node: AnyFlowNode = {
      id: gn.id,
      type: 'function',
      position: { x: 0, y: 0 },
      data: {
        label: gn.name,
        packagePath: gn.package,
        file: gn.file,
        line: gn.line,
        changed: gn.changed,
        clusterId: cid,
        clusterColorHex: color.hex,
        layer,
      },
    }
    const funcDiff = diffData?.files.find(
      (f) => gn.file === f.filename || gn.file.endsWith('/' + f.filename),
    )
    return { panelNode: node, selectedCluster: cluster, selectedLayer: layer, panelDiff: funcDiff }
  }, [data, diffData, selectedNodeId])

  if (isLoading) {
    return (
      <div className="flex h-svh items-center justify-center bg-[#fafaf9]">
        <LoadingSpinner />
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="flex h-svh items-center justify-center bg-[#fafaf9]">
        <Card className="w-full max-w-md border-red-200">
          <CardContent className="space-y-3 pt-6">
            <p className="text-sm font-semibold text-red-600">グラフの読み込みに失敗しました</p>
            <p className="text-xs text-stone-500">
              {error instanceof Error ? error.message : '不明なエラー'}
            </p>
            <Button asChild variant="outline" size="sm">
              <Link to="/">ホームへ戻る</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <AppShell
      topBar={<PRMetaBar pr={data.pr} />}
      rightPanel={
        panelNode ? (
          <FunctionDetailsPanel
            node={panelNode}
            cluster={selectedCluster}
            layer={selectedLayer}
            diff={panelDiff}
            onClose={() => setSelectedNodeId(null)}
          />
        ) : undefined
      }
    >
      <DependencyGraph
        data={data}
        nodeKind={nodeKind}
        onChangeNodeKind={handleChangeNodeKind}
        selectedNodeId={selectedNodeId}
        onSelectNode={setSelectedNodeId}
      />
    </AppShell>
  )
}

function LoadingSpinner() {
  return (
    <div className="flex items-center gap-3 text-stone-400">
      <svg
        width="20"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        className="animate-spin"
      >
        <circle cx="10" cy="10" r="7" strokeDasharray="32" strokeDashoffset="8" />
      </svg>
      <span className="text-sm">グラフを読み込み中…</span>
    </div>
  )
}
