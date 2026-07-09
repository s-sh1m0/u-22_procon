import { useMemo, useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import type { LayerKind, AnyFlowNode } from '@/types/graph'
import type { Cluster, ClusterMode, DiffFile, GraphResponse } from '@/types/api'
import { useGraph } from '@/hooks/useGraph'
import { useDiff } from '@/hooks/useDiff'
import { inferLayer } from '@/lib/layerInference'
import { getClusterColor } from '@/lib/clusterColors'
import { makeClusterKey } from '@/lib/graphLayout'
import { filterToImpact } from '@/lib/impactFilter'
import { isChanged } from '@/lib/nodeChange'
import {
  computeInDegree,
  computeReviewPriority,
  newCycleNodeIds,
  type ReviewPriority,
} from '@/lib/reviewPriority'
import AppShell from '@/components/layout/AppShell'
import PRMetaBar from '@/components/layout/PRMetaBar'
import DependencyGraph from '@/components/graph/DependencyGraph'
import type { SearchCandidate } from '@/components/graph/GraphSearch'
import FunctionDetailsPanel from '@/components/graph/FunctionDetailsPanel'
import CycleAlert from '@/components/graph/CycleAlert'
import LayeringAlert from '@/components/graph/LayeringAlert'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

type Props = { jobId: string }

// デフォルト（初回表示）で自動展開する関数ノード数の予算。これを超える分の
// 変更クラスタは折りたたんだまま開始し、大規模 PR で初期描画が重くなるのを防ぐ。
// ユーザーは個別展開・すべて展開で随時開ける（展開時も仮想化で描画は軽い）。
const AUTO_EXPAND_NODE_BUDGET = 300

export default function AnalysisGraphView({ jobId }: Props) {
  const [clusterMode, setClusterMode] = useState<ClusterMode>('louvain')
  const { data, isLoading, error } = useGraph(jobId, true, clusterMode)
  const { data: diffData } = useDiff(jobId, !isLoading && !error && !!data)
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  // null = 未操作（自動展開ロジックを使う）、Set = ユーザー操作後の明示的な展開セット
  const [expandedClustersOverride, setExpandedClustersOverride] = useState<Set<string> | null>(null)
  // true = 変更影響のみ（diff 1-hop 近傍に絞り込み）、false = 全体グラフ
  const [impactOnly, setImpactOnly] = useState(true)
  // 検索ジャンプで中央寄せしたいノード。nonce は同一ノードへの再ジャンプでも
  // グラフ側の effect を再発火させるための世代番号。
  const [centerTarget, setCenterTarget] = useState<{ nodeId: string; nonce: number } | null>(null)
  const [changedNodeIndex, setChangedNodeIndex] = useState<number | null>(null)

  // 描画に使うグラフ。変更影響ビューでは diff 近傍だけに絞った部分グラフを渡す。
  // クラスタ集計・展開・レイアウトはすべてこの view を基準に行う。
  const view = useMemo(() => {
    if (!data) return data
    return impactOnly ? filterToImpact(data) : data
  }, [data, impactOnly])

  // モード切替時はクラスタが組み変わるため、選択 / 明示的な展開セットをリセットして
  // 新しいクラスタ構造に応じた自動展開ロジックを再適用する。
  const handleChangeClusterMode = useCallback((mode: ClusterMode) => {
    setClusterMode(mode)
    setSelectedNodeId(null)
    setExpandedClustersOverride(null)
    setChangedNodeIndex(null)
  }, [])

  // 影響ビューの切替もクラスタ構成が変わるため、選択 / 展開セットをリセットする。
  const handleSetImpactOnly = useCallback(
    (next: boolean) => {
      if (next === impactOnly) return
      setImpactOnly(next)
      setSelectedNodeId(null)
      setExpandedClustersOverride(null)
      setChangedNodeIndex(null)
    },
    [impactOnly],
  )

  const { defaultExpandedClusters, allClusterKeys } = useMemo(() => {
    const defaultExpanded = new Set<string>()
    const allKeys = new Set<string>()
    if (!view) return { defaultExpandedClusters: defaultExpanded, allClusterKeys: allKeys }

    const nodeMap = new Map(view.graph.nodes.map((n) => [n.id, n]))

    // 展開キー(layer:clusterId)ごとにノード数・変更ノード数を集計する。
    const keyStats = new Map<string, { nodeCount: number; changedCount: number }>()
    for (const cluster of view.clusters) {
      for (const nid of cluster.nodes) {
        const n = nodeMap.get(nid)
        if (!n) continue
        const key = makeClusterKey(n.package, cluster.id)
        allKeys.add(key)
        const s = keyStats.get(key) ?? { nodeCount: 0, changedCount: 0 }
        s.nodeCount++
        if (isChanged(n)) s.changedCount++
        keyStats.set(key, s)
      }
    }

    // 変更を含むキーを優先度順（変更ノード数 desc → ノード数 asc）に並べ、
    // 描画ノード総数が予算を超えない範囲で自動展開する。
    const candidates = [...keyStats.entries()]
      .filter(([, s]) => s.changedCount > 0)
      .sort((a, b) => b[1].changedCount - a[1].changedCount || a[1].nodeCount - b[1].nodeCount)

    let used = 0
    for (const [key, s] of candidates) {
      // 最優先の1キーは予算超過でも必ず開く（常に何か展開された状態で見せる）。
      if (used > 0 && used + s.nodeCount > AUTO_EXPAND_NODE_BUDGET) continue
      defaultExpanded.add(key)
      used += s.nodeCount
    }

    return { defaultExpandedClusters: defaultExpanded, allClusterKeys: allKeys }
  }, [view])

  const expandedClusters = expandedClustersOverride ?? defaultExpandedClusters

  const handleToggleCluster = useCallback(
    (key: string) => {
      setExpandedClustersOverride((prev) => {
        const base = prev !== null ? prev : defaultExpandedClusters
        const next = new Set(base)
        if (next.has(key)) next.delete(key)
        else next.add(key)
        return next
      })
    },
    [defaultExpandedClusters],
  )

  const handleExpandAll = useCallback(() => {
    setExpandedClustersOverride(new Set(allClusterKeys))
  }, [allClusterKeys])

  const handleCollapseAll = useCallback(() => {
    setExpandedClustersOverride(new Set())
    setSelectedNodeId(null)
  }, [])

  // 指定ノードを可視化する: 折りたたみ中なら親クラスタを展開し、そのノードを選択する。
  // source は展開判定に使うグラフ（全体 = data / 影響ビュー = view）。
  const revealNode = useCallback(
    (source: GraphResponse, nodeId: string) => {
      const target = source.graph.nodes.find((n) => n.id === nodeId)
      if (!target) return
      const cluster = source.clusters.find((c) => c.nodes.includes(nodeId))
      if (cluster) {
        const key = makeClusterKey(target.package, cluster.id)
        setExpandedClustersOverride((prev) => {
          const base = prev !== null ? prev : defaultExpandedClusters
          if (base.has(key)) return prev
          const next = new Set(base)
          next.add(key)
          return next
        })
      }
      setSelectedNodeId(nodeId)
    },
    [defaultExpandedClusters],
  )

  // CycleAlert / LayeringAlert からノードを選択された場合: 展開 + 選択 + フォーカス
  const handleSelectCycleNode = useCallback(
    (nodeId: string) => {
      if (data) revealNode(data, nodeId)
    },
    [data, revealNode],
  )

  // 検索ボックスから関数を選択: 現在のビュー内で展開 + 選択し、中央へ寄せる。
  const handleJumpToNode = useCallback(
    (nodeId: string) => {
      if (!view) return
      revealNode(view, nodeId)
      setCenterTarget((t) => ({ nodeId, nonce: (t?.nonce ?? 0) + 1 }))
    },
    [view, revealNode],
  )

  // 検索候補は現在描画しているビュー（影響のみ / 全体）の全関数ノード。
  const searchCandidates = useMemo<SearchCandidate[]>(() => {
    if (!view) return []
    return view.graph.nodes.map((n) => ({ id: n.id, name: n.name, package: n.package }))
  }, [view])

  // in-degree（被呼び出し数）と新規循環ノード集合はグラフ全体から一度だけ算出する。
  const { inDegree, cycleNodeIds } = useMemo(
    () => ({
      inDegree: data ? computeInDegree(data.graph.edges) : new Map<string, number>(),
      cycleNodeIds: data ? newCycleNodeIds(data.cycles) : new Set<string>(),
    }),
    [data],
  )

  // 変更ノードをレビュー優先度順にソートしたリスト。ステップナビゲーションの巡回対象。
  const changedNodeIds = useMemo(() => {
    if (!view) return []
    const priorityOrder: Record<ReviewPriority, number> = { high: 0, medium: 1, low: 2 }
    return view.graph.nodes
      .filter((n) => isChanged(n))
      .map((n) => ({
        id: n.id,
        pkg: n.package,
        priority: computeReviewPriority(n, inDegree.get(n.id) ?? 0, cycleNodeIds.has(n.id)),
      }))
      .sort(
        (a, b) =>
          priorityOrder[a.priority] - priorityOrder[b.priority] || a.pkg.localeCompare(b.pkg),
      )
      .map((n) => n.id)
  }, [view, inDegree, cycleNodeIds])

  const handleNextChanged = useCallback(() => {
    if (changedNodeIds.length === 0 || !view) return
    const next = changedNodeIndex === null ? 0 : (changedNodeIndex + 1) % changedNodeIds.length
    setChangedNodeIndex(next)
    revealNode(view, changedNodeIds[next])
    setCenterTarget((t) => ({ nodeId: changedNodeIds[next], nonce: (t?.nonce ?? 0) + 1 }))
  }, [changedNodeIds, changedNodeIndex, view, revealNode])

  const handlePrevChanged = useCallback(() => {
    if (changedNodeIds.length === 0 || !view) return
    const prev =
      changedNodeIndex === null
        ? changedNodeIds.length - 1
        : (changedNodeIndex - 1 + changedNodeIds.length) % changedNodeIds.length
    setChangedNodeIndex(prev)
    revealNode(view, changedNodeIds[prev])
    setCenterTarget((t) => ({ nodeId: changedNodeIds[prev], nonce: (t?.nonce ?? 0) + 1 }))
  }, [changedNodeIds, changedNodeIndex, view, revealNode])

  const { panelNode, selectedCluster, selectedLayer, panelDiff, panelPriority } = useMemo(() => {
    const empty = {
      panelNode: null,
      selectedCluster: null,
      selectedLayer: 'other' as LayerKind,
      panelDiff: undefined as DiffFile | undefined,
      panelPriority: 'low' as ReviewPriority,
    }
    if (!data || !selectedNodeId) return empty

    const gn = data.graph.nodes.find((n) => n.id === selectedNodeId)
    if (!gn) return empty

    const layer = inferLayer(gn.package)
    const cluster: Cluster | null = data.clusters.find((c) => c.nodes.includes(gn.id)) ?? null
    const cid = cluster?.id ?? 0
    const color = getClusterColor(cid)
    const inCycle = cycleNodeIds.has(gn.id)
    const priority = computeReviewPriority(gn, inDegree.get(gn.id) ?? 0, inCycle)

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
        diffStatus: gn.diff_status,
        inCycle,
        clusterId: cid,
        clusterColorHex: color.hex,
        layer,
      },
    }
    const funcDiff = diffData?.files.find(
      (f) => gn.file === f.filename || gn.file.endsWith('/' + f.filename),
    )
    return {
      panelNode: node,
      selectedCluster: cluster,
      selectedLayer: layer,
      panelDiff: funcDiff,
      panelPriority: priority,
    }
  }, [data, diffData, selectedNodeId, inDegree, cycleNodeIds])

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
      topBar={
        <div className="flex flex-col gap-1">
          <PRMetaBar pr={data.pr} />
          <CycleAlert cycles={data.cycles} onSelectCycleNode={handleSelectCycleNode} />
          <LayeringAlert violations={data.violations} onSelectNode={handleSelectCycleNode} />
        </div>
      }
      rightPanel={
        panelNode ? (
          <FunctionDetailsPanel
            node={panelNode}
            cluster={selectedCluster}
            layer={selectedLayer}
            diff={panelDiff}
            pr={data.pr}
            priority={panelPriority}
            onClose={() => setSelectedNodeId(null)}
          />
        ) : undefined
      }
    >
      <DependencyGraph
        data={view ?? data}
        clusterMode={clusterMode}
        onChangeClusterMode={handleChangeClusterMode}
        impactOnly={impactOnly}
        onSetImpactOnly={handleSetImpactOnly}
        selectedNodeId={selectedNodeId}
        onSelectNode={setSelectedNodeId}
        onClearSelection={() => setSelectedNodeId(null)}
        expandedClusters={expandedClusters}
        onToggleCluster={handleToggleCluster}
        onExpandAll={handleExpandAll}
        onCollapseAll={handleCollapseAll}
        searchCandidates={searchCandidates}
        onJump={handleJumpToNode}
        centerTarget={centerTarget}
        changedCount={changedNodeIds.length}
        changedIndex={changedNodeIndex}
        onNextChanged={handleNextChanged}
        onPrevChanged={handlePrevChanged}
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
