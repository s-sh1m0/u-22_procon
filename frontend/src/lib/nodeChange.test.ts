import { describe, it, expect } from 'vitest'
import { isChanged } from './nodeChange'

describe('isChanged', () => {
  it('changed=true なら変更扱い', () => {
    expect(isChanged({ changed: true, diff_status: 'existing' })).toBe(true)
  })

  it('diff_status が existing 以外なら変更扱い', () => {
    expect(isChanged({ changed: false, diff_status: 'added' })).toBe(true)
    expect(isChanged({ changed: false, diff_status: 'removed' })).toBe(true)
  })

  it('changed=false かつ diff_status=existing なら未変更', () => {
    expect(isChanged({ changed: false, diff_status: 'existing' })).toBe(false)
  })
})
