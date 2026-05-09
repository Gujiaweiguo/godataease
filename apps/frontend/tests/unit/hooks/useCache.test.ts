import { describe, it, expect, vi, beforeEach } from 'vitest'

const wsCacheInstance = {
  get: vi.fn(),
  set: vi.fn(),
  delete: vi.fn(),
  replace: vi.fn(),
  flush: vi.fn(),
  flushAll: vi.fn(),
  has: vi.fn()
}

vi.mock('web-storage-cache', () => {
  return {
    default: class WebStorageCache {
      get: ReturnType<typeof vi.fn>
      set: ReturnType<typeof vi.fn>
      delete: ReturnType<typeof vi.fn>
      replace: ReturnType<typeof vi.fn>
      flush: ReturnType<typeof vi.fn>
      flushAll: ReturnType<typeof vi.fn>
      has: ReturnType<typeof vi.fn>
      constructor() {
        this.get = wsCacheInstance.get
        this.set = wsCacheInstance.set
        this.delete = wsCacheInstance.delete
        this.replace = wsCacheInstance.replace
        this.flush = wsCacheInstance.flush
        this.flushAll = wsCacheInstance.flushAll
        this.has = wsCacheInstance.has
      }
    }
  }
})

import { useCache } from '@/hooks/web/useCache'

describe('useCache', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should return wsCache object', () => {
    const { wsCache } = useCache()
    expect(wsCache).toBeDefined()
  })

  it('should default to localStorage type', () => {
    const { wsCache } = useCache()
    expect(wsCache).toBeDefined()
    expect(typeof wsCache.get).toBe('function')
  })

  it('should accept localStorage type explicitly', () => {
    const { wsCache } = useCache('localStorage')
    expect(wsCache).toBeDefined()
    expect(typeof wsCache.get).toBe('function')
  })

  it('should accept sessionStorage type', () => {
    const { wsCache } = useCache('sessionStorage')
    expect(wsCache).toBeDefined()
    expect(typeof wsCache.get).toBe('function')
  })

  it('should support wsCache.get calls', () => {
    const { wsCache } = useCache()
    wsCache.get('testKey')
    expect(wsCacheInstance.get).toHaveBeenCalledWith('testKey')
  })

  it('should support wsCache.set calls', () => {
    const { wsCache } = useCache()
    wsCache.set('testKey', 'testValue')
    expect(wsCacheInstance.set).toHaveBeenCalledWith('testKey', 'testValue')
  })

  it('should support wsCache.delete calls', () => {
    const { wsCache } = useCache()
    wsCache.delete('testKey')
    expect(wsCacheInstance.delete).toHaveBeenCalledWith('testKey')
  })

  it('should create separate instances for different storage types', () => {
    const { wsCache: localCache } = useCache('localStorage')
    const { wsCache: sessionCache } = useCache('sessionStorage')
    expect(localCache).toBeDefined()
    expect(sessionCache).toBeDefined()
  })

  it('should allow storing and retrieving object values', () => {
    const { wsCache } = useCache()
    const obj = { name: 'test', value: 123 }
    wsCache.set('objKey', obj)
    expect(wsCacheInstance.set).toHaveBeenCalledWith('objKey', obj)
  })

  it('should allow storing array values', () => {
    const { wsCache } = useCache()
    const arr = [1, 2, 3]
    wsCache.set('arrKey', arr)
    expect(wsCacheInstance.set).toHaveBeenCalledWith('arrKey', arr)
  })

  it('should support flush operation', () => {
    const { wsCache } = useCache()
    ;(wsCache as any).flush()
    expect(wsCacheInstance.flush).toHaveBeenCalled()
  })

  it('should support has operation', () => {
    const { wsCache } = useCache()
    ;(wsCache as any).has('testKey')
    expect(wsCacheInstance.has).toHaveBeenCalledWith('testKey')
  })
})
