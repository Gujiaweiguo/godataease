import { describe, it, expect, beforeEach } from 'vitest'
import { clearCache } from '@/utils/cacheUtil'

describe('clearCache', () => {
  const expectedKeys = [
    'DataEaseKey',
    'TreeSort-backend',
    'app.desktop',
    'de-global-refresh',
    'open-backend',
    'panel-weight',
    'screen-weight',
    'user.exp',
    'user.language',
    'user.name',
    'user.oid',
    'user.time',
    'user.token',
    'user.uid'
  ]

  beforeEach(() => {
    localStorage.clear()
  })

  it('should remove all 14 expected keys from localStorage', () => {
    expectedKeys.forEach(key => localStorage.setItem(key, 'some-value'))
    clearCache()
    expectedKeys.forEach(key => {
      expect(localStorage.getItem(key)).toBeNull()
    })
  })

  it('should not remove keys that are not in the clear list', () => {
    localStorage.setItem('unrelated-key', 'keep-me')
    clearCache()
    expect(localStorage.getItem('unrelated-key')).toBe('keep-me')
  })

  it('should handle already-empty localStorage without errors', () => {
    expect(() => clearCache()).not.toThrow()
  })

  it('should remove keys even when only some are set', () => {
    localStorage.setItem('user.token', 'abc123')
    localStorage.setItem('user.name', 'admin')
    clearCache()
    expect(localStorage.getItem('user.token')).toBeNull()
    expect(localStorage.getItem('user.name')).toBeNull()
  })

  it('should not affect keys added after clearing', () => {
    clearCache()
    localStorage.setItem('user.token', 'new-token')
    expect(localStorage.getItem('user.token')).toBe('new-token')
  })

  it('should remove DataEaseKey', () => {
    localStorage.setItem('DataEaseKey', 'value')
    clearCache()
    expect(localStorage.getItem('DataEaseKey')).toBeNull()
  })

  it('should remove all user-prefixed keys', () => {
    const userKeys = expectedKeys.filter(k => k.startsWith('user.'))
    userKeys.forEach(key => localStorage.setItem(key, 'val'))
    clearCache()
    userKeys.forEach(key => {
      expect(localStorage.getItem(key)).toBeNull()
    })
  })

  it('should remove weight-related keys', () => {
    localStorage.setItem('panel-weight', '10')
    localStorage.setItem('screen-weight', '20')
    clearCache()
    expect(localStorage.getItem('panel-weight')).toBeNull()
    expect(localStorage.getItem('screen-weight')).toBeNull()
  })
})
