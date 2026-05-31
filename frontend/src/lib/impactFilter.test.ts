import { describe, it, expect } from 'vitest'
import type { GraphResponse, GraphNode, DiffStatus } from '@/types/api'
import { filterToImpact } from './impactFilter'

function node(id: string, opts: Partial<GraphNode> = {}): GraphNode {
  return {
    id,
    name: id,
    package: 'pkg',
    file: `${id}.go`,
    line: 1,
    changed: false,
    diff_status: 'existing',
    ...opts,
  }
}

function build(overrides: Partial<GraphResponse>): GraphResponse {
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

const edge = (from: string, to: string, status: DiffStatus = 'existing') => ({ from, to, status })

describe('filterToImpact', () => {
  it('変更ノードと直接近傍 1-hop（両方向）だけ残し、無関係なノード/エッジを落とす', () => {
    const data = build({
      graph: {
        // B → A（A の caller）、A → C（A の callee）は残す。
        // D → E（無関係）と B → D（変更ノードに触れない）は落とす。
        nodes: [node('A', { changed: true }), node('B'), node('C'), node('D'), node('E')],
        edges: [edge('B', 'A'), edge('A', 'C'), edge('B', 'D'), edge('D', 'E')],
      },
    })

    const out = filterToImpact(data)

    expect(new Set(out.graph.nodes.map((n) => n.id))).toEqual(new Set(['A', 'B', 'C']))
    expect(out.graph.edges).toEqual([edge('B', 'A'), edge('A', 'C')])
  })

  it('変更ノードに触れないエッジ経由でしか繋がらないノード（近傍の近傍）は落とす', () => {
    const data = build({
      graph: {
        nodes: [node('A', { changed: true }), node('B'), node('D')],
        edges: [edge('B', 'A'), edge('B', 'D')],
      },
    })

    const out = filterToImpact(data)

    // B は A の caller なので残るが、D は変更ノードに触れないため残さない。
    expect(new Set(out.graph.nodes.map((n) => n.id))).toEqual(new Set(['A', 'B']))
    expect(out.graph.edges).toEqual([edge('B', 'A')])
  })

  it('diff_status が added / removed のノードも変更扱いで残す（孤立していても残す）', () => {
    const data = build({
      graph: {
        nodes: [
          node('F', { diff_status: 'added' }),
          node('G', { diff_status: 'removed' }),
          node('H'),
        ],
        edges: [],
      },
    })

    const out = filterToImpact(data)

    expect(new Set(out.graph.nodes.map((n) => n.id))).toEqual(new Set(['F', 'G']))
    expect(out.graph.edges).toEqual([])
  })

  it('clusters は残ったノードで絞り込み、空になったクラスタは除去する', () => {
    const data = build({
      graph: {
        nodes: [node('A', { changed: true }), node('C'), node('D'), node('E')],
        edges: [edge('A', 'C'), edge('D', 'E')],
      },
      clusters: [
        { id: 0, label: 'c0', nodes: ['A', 'C'] },
        { id: 1, label: 'c1', nodes: ['D', 'E'] },
      ],
    })

    const out = filterToImpact(data)

    expect(out.clusters).toEqual([{ id: 0, label: 'c0', nodes: ['A', 'C'] }])
  })

  it('入力を破壊しない', () => {
    const data = build({
      graph: {
        nodes: [node('A', { changed: true }), node('D')],
        edges: [edge('A', 'A')],
      },
      clusters: [{ id: 0, label: 'c0', nodes: ['A', 'D'] }],
    })

    filterToImpact(data)

    expect(data.graph.nodes.map((n) => n.id)).toEqual(['A', 'D'])
    expect(data.clusters[0].nodes).toEqual(['A', 'D'])
  })
})
