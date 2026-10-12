import { describe, it, expect } from 'vitest'
import { collapseImplementations, implementationIds } from './interfaceCollapse'
import {
  testNode as node,
  testIfaceNode as iface,
  testEdge as edge,
  testImplEdge as impl,
  testGraphResponse as build,
} from './testGraph'

const ids = (data: { graph: { nodes: { id: string }[] } }) =>
  new Set(data.graph.nodes.map((n) => n.id))

describe('collapseImplementations', () => {
  const graph = {
    nodes: [
      node('S', { changed: true }),
      iface('I'),
      node('Impl1'),
      node('Impl2'),
      node('ImplCallee'),
    ],
    edges: [edge('S', 'I'), impl('I', 'Impl1'), impl('I', 'Impl2'), edge('Impl1', 'ImplCallee')],
  }

  it('展開されていない interface メソッドの実装と implements エッジを隠す', () => {
    const data = build({
      graph,
      clusters: [
        { id: 0, label: 'c0', nodes: ['S', 'I', 'ImplCallee'] },
        { id: 1, label: 'c1', nodes: ['Impl1', 'Impl2'] },
      ],
    })

    const out = collapseImplementations(data, new Set())

    expect(ids(out.data)).toEqual(new Set(['S', 'I', 'ImplCallee']))
    expect(out.data.graph.edges).toEqual([edge('S', 'I')])
    expect(out.data.clusters).toEqual([{ id: 0, label: 'c0', nodes: ['S', 'I', 'ImplCallee'] }])
    expect(out.implCounts.get('I')).toBe(2)
  })

  it('展開された interface メソッドの実装は残す', () => {
    const data = build({ graph })

    const out = collapseImplementations(data, new Set(['I']))

    expect(ids(out.data)).toEqual(ids(data))
    expect(out.data.graph.edges).toEqual(data.graph.edges)
    expect(out.implCounts.get('I')).toBe(2)
  })

  it('変更された実装は折りたたみ中も implements エッジごと残す', () => {
    const data = build({
      graph: {
        nodes: [iface('I'), node('Caller'), node('S', { changed: true }), node('OtherImpl')],
        edges: [edge('Caller', 'I'), impl('I', 'S'), impl('I', 'OtherImpl')],
      },
    })

    const out = collapseImplementations(data, new Set())

    expect(ids(out.data)).toEqual(new Set(['I', 'Caller', 'S']))
    expect(out.data.graph.edges).toEqual([edge('Caller', 'I'), impl('I', 'S')])
  })

  it('直接呼ばれている実装はノードを残し、implements エッジだけ隠す', () => {
    const data = build({
      graph: {
        nodes: [node('S', { changed: true }), iface('I'), node('Impl')],
        edges: [edge('S', 'I'), edge('S', 'Impl'), impl('I', 'Impl')],
      },
    })

    const out = collapseImplementations(data, new Set())

    expect(ids(out.data)).toEqual(new Set(['S', 'I', 'Impl']))
    expect(out.data.graph.edges).toEqual([edge('S', 'I'), edge('S', 'Impl')])
  })

  it('入力を破壊しない', () => {
    const data = build({ graph, clusters: [{ id: 0, label: 'c0', nodes: ['Impl1'] }] })

    collapseImplementations(data, new Set())

    expect(data.graph.nodes).toHaveLength(5)
    expect(data.graph.edges).toHaveLength(4)
    expect(data.clusters[0].nodes).toEqual(['Impl1'])
  })
})

describe('implementationIds', () => {
  it('implements エッジの先を返す', () => {
    const data = build({
      graph: {
        nodes: [iface('I'), iface('J'), node('A'), node('B')],
        edges: [impl('I', 'A'), impl('J', 'B'), edge('A', 'I')],
      },
    })

    expect(implementationIds(data, 'I')).toEqual(['A'])
  })
})
