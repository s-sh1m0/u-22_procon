import { describe, it, expect } from 'vitest'
import { computeInDegree } from './reviewPriority'
import { testEdge as edge, testImplEdge as impl } from './testGraph'

describe('computeInDegree', () => {
  it('呼び出しエッジの呼び出し元を数える', () => {
    const m = computeInDegree([edge('A', 'C'), edge('B', 'C'), edge('C', 'D')])

    expect(m.get('C')).toBe(2)
    expect(m.get('D')).toBe(1)
    expect(m.get('A')).toBeUndefined()
  })

  it('実装には interface メソッドの被呼び出し数を足し、implements エッジ自体は数えない', () => {
    const m = computeInDegree([
      edge('A', 'I'),
      edge('B', 'I'),
      impl('I', 'X'),
      impl('I', 'Y'),
      edge('C', 'X'),
    ])

    expect(m.get('I')).toBe(2)
    expect(m.get('X')).toBe(3) // 直接の呼び出し 1 + interface 経由 2
    expect(m.get('Y')).toBe(2)
  })
})
