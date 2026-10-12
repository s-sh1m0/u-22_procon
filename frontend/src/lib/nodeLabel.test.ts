import { describe, it, expect } from 'vitest'
import { nodeDisplayName } from './nodeLabel'

describe('nodeDisplayName', () => {
  it('メソッドはレシーバ型を前に付ける', () => {
    expect(nodeDisplayName({ name: 'TargetContains', receiver: 'AbsAction' })).toBe(
      'AbsAction.TargetContains',
    )
  })

  it('関数（レシーバなし）は名前だけを返す', () => {
    expect(nodeDisplayName({ name: 'planAction', receiver: '' })).toBe('planAction')
    expect(nodeDisplayName({ name: 'planAction' })).toBe('planAction')
  })
})
