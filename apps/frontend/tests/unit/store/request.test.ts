import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useRequestStore } from '@/store/modules/request'

describe('Request Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have empty loadingMap', () => {
      const store = useRequestStore()
      expect(store.loadingMap).toEqual({})
    })

    it('should have empty cachedRequestList', () => {
      const store = useRequestStore()
      expect(store.cachedRequestList).toEqual([])
    })

    it('should return empty array from getRequestList getter', () => {
      const store = useRequestStore()
      expect(store.getRequestList).toEqual([])
    })
  })

  describe('setLoadingMap', () => {
    it('should set loadingMap to the given value', () => {
      const store = useRequestStore()
      store.setLoadingMap({ 'api/data': 2, 'api/user': 1 })
      expect(store.loadingMap).toEqual({ 'api/data': 2, 'api/user': 1 })
    })

    it('should replace entire loadingMap', () => {
      const store = useRequestStore()
      store.setLoadingMap({ 'api/a': 1 })
      store.setLoadingMap({ 'api/b': 3 })
      expect(store.loadingMap).toEqual({ 'api/b': 3 })
    })
  })

  describe('resetLoadingMap', () => {
    it('should set all loadingMap values to 0', () => {
      const store = useRequestStore()
      store.setLoadingMap({ 'api/a': 3, 'api/b': 5 })
      store.resetLoadingMap()
      expect(store.loadingMap).toEqual({ 'api/a': 0, 'api/b': 0 })
    })

    it('should handle empty loadingMap gracefully', () => {
      const store = useRequestStore()
      store.resetLoadingMap()
      expect(store.loadingMap).toEqual({})
    })
  })

  describe('addLoading', () => {
    it('should create new key with value 1', () => {
      const store = useRequestStore()
      store.addLoading('api/new')
      expect(store.loadingMap['api/new']).toBe(1)
    })

    it('should increment existing key', () => {
      const store = useRequestStore()
      store.setLoadingMap({ 'api/data': 2 })
      store.addLoading('api/data')
      expect(store.loadingMap['api/data']).toBe(3)
    })
  })

  describe('reduceLoading', () => {
    it('should decrement existing key', () => {
      const store = useRequestStore()
      store.setLoadingMap({ 'api/data': 3 })
      store.reduceLoading('api/data')
      expect(store.loadingMap['api/data']).toBe(2)
    })

    it('should decrement to negative if already 0', () => {
      const store = useRequestStore()
      store.setLoadingMap({ 'api/data': 0 })
      store.reduceLoading('api/data')
      expect(store.loadingMap['api/data']).toBe(-1)
    })
  })

  describe('addCacheRequest', () => {
    it('should add a function to cachedRequestList', () => {
      const store = useRequestStore()
      const fn = vi.fn()
      store.addCacheRequest(fn)
      expect(store.cachedRequestList).toHaveLength(1)
      expect(store.getRequestList).toHaveLength(1)
    })

    it('should add multiple functions', () => {
      const store = useRequestStore()
      const fn1 = vi.fn()
      const fn2 = vi.fn()
      store.addCacheRequest(fn1)
      store.addCacheRequest(fn2)
      expect(store.cachedRequestList).toHaveLength(2)
    })

    it('should store functions that can be called', () => {
      const store = useRequestStore()
      const fn = vi.fn()
      store.addCacheRequest(fn)
      store.cachedRequestList[0]('test-token')
      expect(fn).toHaveBeenCalledWith('test-token')
    })
  })

  describe('cleanCacheRequest', () => {
    it('should clear all cached requests', () => {
      const store = useRequestStore()
      store.addCacheRequest(vi.fn())
      store.addCacheRequest(vi.fn())
      store.cleanCacheRequest()
      expect(store.cachedRequestList).toEqual([])
      expect(store.getRequestList).toEqual([])
    })
  })
})
