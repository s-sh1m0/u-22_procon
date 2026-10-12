import type { DiffStatus, GraphEdge, GraphNode, GraphResponse } from '@/types/api'

// グラフを扱う純粋関数のテスト用ヘルパー。

/** 関数ノードを作る。opts で任意のフィールドを上書きする。 */
export function testNode(id: string, opts: Partial<GraphNode> = {}): GraphNode {
  return {
    id,
    kind: 'func',
    name: id,
    receiver: '',
    package: 'pkg',
    file: `${id}.go`,
    line: 1,
    changed: false,
    diff_status: 'existing',
    ...opts,
  }
}

/** interface メソッドノードを作る。 */
export function testIfaceNode(id: string, opts: Partial<GraphNode> = {}): GraphNode {
  return testNode(id, { kind: 'interface_method', ...opts })
}

/** 呼び出しエッジを作る。 */
export function testEdge(from: string, to: string, status: DiffStatus = 'existing'): GraphEdge {
  return { from, to, kind: 'call', status }
}

/** interface メソッドから実装への implements エッジを作る。 */
export function testImplEdge(from: string, to: string, status: DiffStatus = 'existing'): GraphEdge {
  return { from, to, kind: 'implements', status }
}

/** 空の GraphResponse に overrides を重ねて返す。 */
export function testGraphResponse(overrides: Partial<GraphResponse>): GraphResponse {
  return {
    pr: {
      owner: 'o',
      repo: 'r',
      number: 1,
      title: 't',
      base_ref: 'main',
      head_ref: 'feat',
      base_sha: 'b',
      head_sha: 'h',
    },
    clusters: [],
    graph: { nodes: [], edges: [] },
    cycles: [],
    violations: [],
    ...overrides,
  }
}
