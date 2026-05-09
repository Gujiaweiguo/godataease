import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useShareStore } from '@/store/modules/share'

describe('Share Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Initial State', () => {
    it('should have shareDisable as false', () => {
      const store = useShareStore()
      expect(store.shareDisable).toBe(false)
    })

    it('should have sharePeRequire as false', () => {
      const store = useShareStore()
      expect(store.sharePeRequire).toBe(false)
    })

    it('should return false from getShareDisable getter', () => {
      const store = useShareStore()
      expect(store.getShareDisable).toBe(false)
    })

    it('should return false from getSharePeRequire getter', () => {
      const store = useShareStore()
      expect(store.getSharePeRequire).toBe(false)
    })
  })

  describe('setData', () => {
    it('should update shareDisable to true', () => {
      const store = useShareStore()
      store.setData({ shareDisable: true, sharePeRequire: false })
      expect(store.shareDisable).toBe(true)
      expect(store.getShareDisable).toBe(true)
    })

    it('should update sharePeRequire to true', () => {
      const store = useShareStore()
      store.setData({ shareDisable: false, sharePeRequire: true })
      expect(store.sharePeRequire).toBe(true)
      expect(store.getSharePeRequire).toBe(true)
    })

    it('should update both flags simultaneously', () => {
      const store = useShareStore()
      store.setData({ shareDisable: true, sharePeRequire: true })
      expect(store.getShareDisable).toBe(true)
      expect(store.getSharePeRequire).toBe(true)
    })

    it('should reset both flags to false', () => {
      const store = useShareStore()
      store.setData({ shareDisable: true, sharePeRequire: true })
      store.setData({ shareDisable: false, sharePeRequire: false })
      expect(store.getShareDisable).toBe(false)
      expect(store.getSharePeRequire).toBe(false)
    })

    it('should toggle shareDisable independently', () => {
      const store = useShareStore()
      store.setData({ shareDisable: true, sharePeRequire: false })
      expect(store.getShareDisable).toBe(true)
      expect(store.getSharePeRequire).toBe(false)
    })
  })
})
